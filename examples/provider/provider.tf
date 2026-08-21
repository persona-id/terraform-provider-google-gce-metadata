terraform {
  required_providers {
    google-gce-metadata = {
      source = "persona-id/google-gce-metadata"
    }
  }
}

# The provider makes no API calls and takes no configuration.
provider "google-gce-metadata" {}
