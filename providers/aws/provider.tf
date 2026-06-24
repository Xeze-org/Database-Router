# Compute provider — reads AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY from
# the environment. Region is driven by var.region (default us-east-1).
provider "aws" {
  region = var.region
}

# DNS provider — reads CLOUDFLARE_API_TOKEN from the environment.
provider "cloudflare" {}
