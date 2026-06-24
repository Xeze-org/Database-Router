# Compute provider — reads ARM_CLIENT_ID / ARM_CLIENT_SECRET /
# ARM_SUBSCRIPTION_ID / ARM_TENANT_ID from the environment.
provider "azurerm" {
  features {}
}

# DNS provider — reads CLOUDFLARE_API_TOKEN from the environment.
provider "cloudflare" {}
