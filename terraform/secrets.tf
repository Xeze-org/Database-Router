############################################
# Auto-generated passwords
# Shown in `terraform output` after apply.
# Stored in state — never commit tfstate.
############################################

resource "random_password" "postgres" {
  length  = 24
  special = false
}

resource "random_password" "mongo" {
  length  = 24
  special = false
}

resource "random_password" "redis" {
  length  = 24
  special = false
}

