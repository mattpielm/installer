# OpenShift Installer

## Summary of Changes from release-4.14
```diff
diff --git a/data/data/aws/bootstrap/main.tf b/data/data/aws/bootstrap/main.tf
index 7370b697a9..c4e9af3238 100644
--- a/data/data/aws/bootstrap/main.tf
+++ b/data/data/aws/bootstrap/main.tf
@@ -91,6 +91,7 @@ resource "aws_iam_role" "bootstrap" {

   name = "${var.cluster_id}-bootstrap-role"
   path = "/"
+  force_detach_policies = true

   assume_role_policy = <<EOF
 {
diff --git a/data/data/aws/cluster/main.tf b/data/data/aws/cluster/main.tf
index e48df9fcf2..bfb4c98cbb 100644
--- a/data/data/aws/cluster/main.tf
+++ b/data/data/aws/cluster/main.tf
@@ -66,6 +66,7 @@ module "iam" {
   tags = local.tags
 }

+/* NoRoute53: Comment out DNS module so installer doesn't even check for route53 (and fail)
 module "dns" {
   source = "./route53"

@@ -84,6 +85,7 @@ module "dns" {
   publish_strategy         = var.aws_publish_strategy
   custom_endpoints         = var.custom_endpoints
 }
+*/

 module "vpc" {
   source = "./vpc"
diff --git a/data/data/azure/cluster/main.tf b/data/data/azure/cluster/main.tf
index 2a5eb0cec7..ec1b1b39b0 100644
--- a/data/data/azure/cluster/main.tf
+++ b/data/data/azure/cluster/main.tf
@@ -57,6 +57,7 @@ module "master" {
   use_ipv6 = var.use_ipv6
 }

+/* NoDNS NoRoute53
 module "dns" {
   source                          = "./dns"
   cluster_domain                  = var.cluster_domain
@@ -75,3 +76,4 @@ module "dns" {
   use_ipv4 = var.use_ipv4
   use_ipv6 = var.use_ipv6
 }
+End NoRoute53 NoDNS */
```

## Supported Platforms

* [AWS](docs/user/aws/README.md)
* [AWS (UPI)](docs/user/aws/install_upi.md)
* [Azure](docs/user/azure/README.md)
* [Bare Metal (UPI)](docs/user/metal/install_upi.md)
* [Bare Metal (IPI)](docs/user/metal/install_ipi.md)
* [GCP](docs/user/gcp/README.md)
* [GCP (UPI)](docs/user/gcp/install_upi.md)
* [Libvirt with KVM](docs/dev/libvirt/README.md) (development only)
* [OpenStack](docs/user/openstack/README.md)
* [OpenStack (UPI)](docs/user/openstack/install_upi.md)
* [Power](docs/user/power/install_upi.md)
* [oVirt](docs/user/ovirt/install_ipi.md)
* [oVirt (UPI)](docs/user/ovirt/install_upi.md)
* [vSphere](docs/user/vsphere/README.md)
* [vSphere (UPI)](docs/user/vsphere/install_upi.md)
* [z/VM](docs/user/zvm/install_upi.md)

## Quick Start

First, install all [build dependencies](docs/dev/dependencies.md).

Clone this repository. Then build the `openshift-install` binary with:

```sh
hack/build.sh
```

This will create `bin/openshift-install`. This binary can then be invoked to create an OpenShift cluster, like so:

```sh
bin/openshift-install create cluster
```

The installer will show a series of prompts for user-specific information and use reasonable defaults for everything else.
In non-interactive contexts, prompts can be bypassed by [providing an `install-config.yaml`](docs/user/overview.md#multiple-invocations).

If you have trouble, refer to [the troubleshooting guide](docs/user/troubleshooting.md).

### Connect to the cluster

Details for connecting to your new cluster are printed by the `openshift-install` binary upon completion, and are also available in the `.openshift_install.log` file.

Example output:

```sh
INFO Waiting 10m0s for the openshift-console route to be created...
INFO Install complete!
INFO To access the cluster as the system:admin user when using 'oc', run
    export KUBECONFIG=/path/to/installer/auth/kubeconfig
INFO Access the OpenShift web-console here: https://console-openshift-console.apps.${CLUSTER_NAME}.${BASE_DOMAIN}:6443
INFO Login to the console with user: kubeadmin, password: 5char-5char-5char-5char
```

### Cleanup

Destroy the cluster and release associated resources with:

```sh
openshift-install destroy cluster
```

Note that you almost certainly also want to clean up the installer state files too, including `auth/`, `terraform.tfstate`, etc.
The best thing to do is always pass the `--dir` argument to `create` and `destroy`.
And if you want to reinstall from scratch, `rm -rf` the asset directory beforehand.
