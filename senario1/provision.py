# Copyright 2021 Yoshi Yamaguchi
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import argparse
import getpass
import os
from fabric import Connection

GOLANG_DOWNLOAD_URL="https://golang.org/dl/go1.15.7.linux-amd64.tar.gz"

class FabricOptions:
    pass

def cmd(*args):
    return ' '.join(args)

parser = argparse.ArgumentParser(description="values to connect servers")
parser.add_argument("--ip", type=str)
args = parser.parse_args(namespace=FabricOptions)

ssh_pass=getpass.getpass("ssh key pass: ")

c = Connection(
    host=FabricOptions.ip,
    user=getpass.getuser(),
    connect_kwargs={
        "key_filename": os.path.join(os.environ['HOME'], ".ssh", "google_compute_engine"),
        "passphrase": ssh_pass,
    },
)

c.run("hostname")
c.run("cat /etc/os-release")
c.sudo(cmd("apt-get", "install", "-y", "wget"))
c.run(cmd("wget", "-O", "go.tgz", GOLANG_DOWNLOAD_URL))
c.run(cmd("tar", "xf", "go.tgz"))
c.sudo(cmd("mv", "go", "/opt"))
c.sudo(cmd("chmod", "a+x", "/opt/go/bin"))
for f in ["main.go", "go.mod", "go.sum"]:
    c.put(f)
c.run(cmd("/opt/go/bin/go", "build", "-o", "app"))
c.run(cmd("nohup", "./app", "&"))
