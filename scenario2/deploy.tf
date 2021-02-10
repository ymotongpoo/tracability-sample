// Copyright 2021 Yoshi Yamaguchi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

variable "project_id" {
    type = string
    description = "Google Cloud Platform project ID"
}

variable "region" {
    type = string
    description = "Google Compute Engine region"
    default = "asia-east1"
}

variable "zone" {
    type = string
    description = "Google Compute Engine zone"
    default = "asia-east1-b"
}

variable "user" {
    type = string
    description = "username for SSH"
    default = "demo"
}

provider "google" {
    project = var.project_id
    zone = var.zone
    region = var.region
}

resource "google_compute_firewall" "dev_http" {
    name = "dev-http"
    network = "default"

    allow {
        protocol = "tcp"
        ports = ["8080"]
    }

    source_ranges = ["0.0.0.0/0"]
    target_tags = ["tracability-sample"]
}

resource "google_compute_instance" "server" {
    name = "scenario2-server"
    machine_type = "e2-standard-2"
    zone = var.zone
    tags = ["tracability-sample"]

    boot_disk {
        initialize_params {
            image = "debian-cloud/debian-10"
        }
    }

    network_interface {
        network = "default"
        access_config {}
    }

    scheduling {
        automatic_restart = true
    }

    depends_on = [google_compute_firewall.dev_http]

    service_account {
        scopes = ["compute-rw", "logging-write", "monitoring"]
    }

    provisioner "local-exec" {
        working_dir = "./ansible/"
        command = "ansible-playbook -i inventory/hosts.gcp.yaml server.yaml"
    }
}

resource "google_compute_instance" "client" {
    name = "scenario2-client"
    machine_type = "e2-standard-2"
    zone = var.zone
    tags = ["tracability-sample"]

    boot_disk {
        initialize_params {
            image = "debian-cloud/debian-10"
        }
    }

    network_interface {
        network = "default"
        access_config {}
    }

    scheduling {
        automatic_restart = true
    }

    depends_on = [google_compute_firewall.dev_http]

    service_account {
        scopes = ["compute-rw", "logging-write", "monitoring"]
    }

    provisioner "local-exec" {
        working_dir = "./ansible/"
        command = "ansible-playbook -i inventory/hosts.gcp.yaml client.yaml --extra-vars \"zone=${var.zone} project_id=${var.project_id} endpoint=${google_compute_instance.server.network_interface.0.access_config.0.nat_ip}\""
    }
}