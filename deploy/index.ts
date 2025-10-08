import {
  Deployment,
  externalIngressAnnotations,
  getEnv,
  hostnamePrefix,
  internalIngressAnnotations,
  k8sProvider,
} from "@jedrw/tk3s-deployment";
import * as pulumi from "@pulumi/pulumi";
import * as doppler from "@pulumiverse/doppler";
import * as yaml from "js-yaml";

export = async () => {
  const appName = pulumi.getProject();
  const httpHostname = `${hostnamePrefix()}${appName}.lupinelab.co.uk`;
  const sshHostname = `${hostnamePrefix()}manage-${appName}.lupinelab.co.uk`;
  const expose = getEnv(true) == "production" ? "external" : "internal";
  const releaseName = `${getEnv()}-${appName}`;
  const contentDir =
    getEnv(true) === "production"
      ? "/tcdata/nfs/brain"
      : "/tcdata/nfs/dev-brain";

  const secrets = await doppler.getSecrets({
    project: "brain",
    config: getEnv(),
  });

  const mkdocsConfig = {
    site_name: "Brainfiles",
    theme: {
      name: "material",
    },
    watch: ["docs"],
    not_in_nav: `/index.md`,
  };

  new Deployment(
    appName,
    {
      hostname: httpHostname,
      expose,
      name: releaseName,
      chart: `../chart/${appName}`,
      namespace: releaseName,
      createNamespace: true,
      values: {
        config: {
          mkdocsConfig: yaml.dump(mkdocsConfig),
          authorizedKeys: [
            "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCmwrnNvwh6T8JcUJfAZT+12grJoP1o7JMIgDsudsI/nEKGrie6fTZ5+bhpfN9dNunnf5xALOfs4dExzGimsNZL7dEp+0pJuSM4NZ36z6JmWwJAojyhGG2E/5hK5oY1BrmZFxYk7komlLlyE7Ypdse/F5Chqw5a5X9aOYQqdlEeMN0YyDsujJ9cnKpYOmM8wdtXNFyg7uOrfWJQVfgVJCY0K5LsOV3uH6nRNIhKOmvbCMjXf99W3xib/ByHQXmWTIsOwtR5qCJDy6aOvoSKIxiBgRYcmfgylHyPV2YLlMKPT0hij5zyRR6jpOBKpoD3w5BAwBRyGpPoIdzZAXg1NkFLEIbYgk8kyBR8IMYvGp+AI58sQK4hSbVuESE5oWOzqLr6aikPjYjdPWUooq/N2G4yd16daM2+rpTO6H7YbjDVfeI4NI5WiVb3yQ8dVQwkhcMX5MVwKGWlypo+EdajEE8Bk1bmhLDpXfDdtg3XZvwa9flQFIpA92TtkVrZn74FpIs=",
          ],
          hostPublicKey:
            "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIGGel5Jn4yne/cCGij1yZTVpVrMfApSZtFKqEz2qJ1wv",
        },
        secrets: {
          hostKey: secrets.map["HOSTKEY"],
        },
        image: {
          tag:
            process.env[`${appName.toUpperCase()}_VERSION`] ||
            process.env["CIRCLE_SHA1"],
        },
        persistence: {
          nfs: {
            server: "tshare.lupinelab",
            path: contentDir,
          },
        },
        ingress: {
          http: {
            annotations:
              getEnv(true) == "production"
                ? {
                    ...externalIngressAnnotations({ hostname: httpHostname }),
                    // Add pfsense DNS entry to resolve directly to MetalLB IP
                    // internally to avoid NAT reflection issues
                    "dns.pfsense.org/enabled": "true",
                  }
                : internalIngressAnnotations(),
            hosts: [
              {
                host: httpHostname,
                paths: [
                  {
                    path: "/",
                    pathType: "ImplementationSpecific",
                  },
                ],
              },
              {
                host: sshHostname,
                paths: [
                  {
                    path: "/",
                    pathType: "ImplementationSpecific",
                  },
                ],
              },
            ],
            tls: [
              {
                hosts: [httpHostname],
                secretName: `${httpHostname}-cert`,
              },
            ],
          },
          ssh: {
            annotations:
              getEnv(true) == "production"
                ? {
                    ...externalIngressAnnotations({ hostname: sshHostname }),
                    // Cloudflare does TLS on proxied stuff so this breaks SSH
                    // connections.
                    "external-dns.alpha.kubernetes.io/cloudflare-proxied":
                      "false",
                    // Not required as we do not do TLS termination for this
                    // ingress and we use a different entrypoint to normal
                    "cert-manager.io/cluster-issuer": undefined,
                    "traefik.ingress.kubernetes.io/router.entrypoints":
                      undefined,
                  }
                : {},
          },
        },
      },
    },
    {
      provider: k8sProvider(),
    }
  );
};
