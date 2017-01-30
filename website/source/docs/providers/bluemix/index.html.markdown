---
layout: "bluemix"
page_title: "Provider: Bluemix"
sidebar_current: "docs-bluemix-index"
description: |-
  The Bluemix provider is used to interact with Bluemix resources.
---

# Bluemix Provider

The Bluemix provider is used to manage Bluemix resources.

Use the navigation to the left to read about the available resources.

-> **Note:** The Bluemix provider is new as of Terraform x.x.x.
It is ready to be used but many features are still being added. If there
is a Bluemix feature missing, please report it in the GitHub repo.

## Example Usage

Here is an example that will setup the following:

+ An SSH key resource.
+ A virtual server resource that uses an existing SSH key.
+ A virtual server resource using an existing SSH key and a Terraform managed SSH key (created as `test_key_1` in the example below).

Add the below to a file called `sl.tf` and run the `terraform` command from the same directory:

```hcl
provider "bluemix" {
    username = ""
    password = ""
    softlayer_username = ""
    softlayer_api_key = ""
}

# This will create a new SSH key that will show up under the \
# Devices>Manage>SSH Keys in the SoftLayer console.
resource "bluemix_infrastructure_ssh_key" "test_key_1" {
    name = "test_key_1"
    public_key = "${file(\"~/.ssh/id_rsa_test_key_1.pub\")}"
    # Windows Example:
    # public_key = "${file(\"C:\ssh\keys\path\id_rsa_test_key_1.pub\")}"
}

# Virtual Server created with existing SSH Key already in SoftLayer \
# inventory and not created using this Terraform template.
resource "bluemix_infrastructure_virtual_guest" "my_server_1" {
    name = "my_server_1"
    domain = "example.com"
    ssh_keys = ["123456"]
    image = "DEBIAN_7_64"
    region = "ams01"
    public_network_speed = 10
    cpu = 1
    ram = 1024
}

# Virtual Server created with a mix of previously existing and \
# Terraform created/managed resources.
resource "bluemix_infrastructure_virtual_guest" "my_server_2" {
    name = "my_server_2"
    domain = "example.com"
    ssh_keys = ["123456", "${bluemix_infrastructure_ssh_key.test_key_1.id}"]
    image = "CENTOS_6_64"
    region = "ams01"
    public_network_speed = 10
    cpu = 1
    ram = 1024
}
```

You'll need to provide your Bluemix username and password,
as well as SoftLayer username and API key,so that Terraform can connect. 
If you don't want to put credentials in your configuration file, you can leave them
out:

```
provider "bluemix" {}
```

...and instead set these environment variables:

- **BLUEMIX_USERNAME/BM_USERNAME**: Your Bluemix username
- **BLUEMIX_PASSWORD/BM_PASSWORD**: Your Bluemix password
- **SOFTLAYER_USERNAME/SL_USERNAME**: Your SoftLayer username
- **SOFTLAYER_API_KEY/SL_API_KEY**: Your SoftLayer API key
