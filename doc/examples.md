# Getting Started with the Examples

`fablab` currently ships with a number of example models. These models all use Fedora 30 as a runtime environment, and will require that your `ZITI_ROOT` contain a Ziti build based on a Fedora 30-compatible `golang` environment. I use Fedora 30 as my development environment, and this is useful because it allows me to `go install` locally, and then `fablab sync` to push changes into my environment. You don't necessarily need to be running Fedora 30 to try the examples, but you will need a compatible Ziti build.

## Requirements

You'll need a Fedora 30-compatible build of Ziti, as mentioned above. Place this in `ZITI_ROOT`. `ZITI_ROOT/bin` will ultimately serve as the `bin` tree that is `sync`ed onto the environment.

`fablab` (currently) has external dependencies on `terraform` and `rsync`. You'll need to have both of these executables available in your shell's `PATH`.

The example models work against AWS. You'll need to have access to AWS credentials, which allow you to create instances in multiple regions, create VPCs and all of the related networking infrastructure. Your AWS account will need an SSH key set up that you can use, and you'll want to have that key loaded into your SSH agent. With all of that together, you'll create a `~/.fablab/user.yml` file, structured like this:

```
environment_tag: "republic_of_q"
aws_access_key: "<AWS Access Key>"
aws_secret_key: "<AWS Secret Key>"
aws_key_path: "/home/michael/.ssh/nf-fablab-michael"
aws_key_name: "mquigley"
```

The `enviroment_tag` defines a name, which makes finding your resources easier using AWS tooling. The AWS keys should be self-evident. The `aws_key_path` and `aws_key_name` refer to the SSH key for the instances you will be creating in AWS.

Once your `user.yml` is together, you'll need to set the following environment variables:

* `ZITI_ROOT`, which should be pointing at the `GOPATH` for your `ziti` repository. It should be built, and contain the build that you'd like to replicate onto your environment.
* `FABLAB_ROOT`, which should be pointing at the root of your `fablab` clone.
* `FABLAB_RUN`, which should be pointing at the location where you want to store your current run kit, storing the state for your environment.

## Create the Run Kit

First, we'll create the run kit:

```
$ fablab create diamond
[   0.000]    INFO fablab/kernel.resolveRunPath: resolved run path [/home/michael/Repos/nf/fablab/build/run]
[   0.000] WARNING fablab/kernel.bootstrapLabel: no run label at run path [/home/michael/Repos/nf/fablab/build/run]
[   0.000] WARNING fablab/kernel.bootstrapBinding: no run label found
[   0.001]    INFO fablab/cmd/fablab/subcmd.create: created run for model [diamond]
```

This will initialize our run kit, setting it up to work with the `diamond` model. 

## Express the Infrastructure

Next, we'll want to apply the infrastructure expressed by `fablab`:

```
$ fablab express
   0.000]    INFO fablab/kernel.resolveRunPath: resolved run path [/home/michael/Repos/nf/fablab/build/run]
[   0.000] WARNING fablab/kernel.bootstrapBinding: no binding [initiator_host_ctrl_public_ip]
[   0.000] WARNING fablab/kernel.bootstrapBinding: no binding [initiator_host_ctrl_private_ip]
[   0.000] WARNING fablab/kernel.bootstrapBinding: no binding [initiator_host_001_public_ip]
[   0.000] WARNING fablab/kernel.bootstrapBinding: no binding [initiator_host_001_private_ip]
[   0.000] WARNING fablab/kernel.bootstrapBinding: no binding [transitA_host_002_public_ip]
[   0.001] WARNING fablab/kernel.bootstrapBinding: no binding [transitA_host_002_private_ip]
[   0.001] WARNING fablab/kernel.bootstrapBinding: no binding [transitB_host_004_public_ip]
[   0.001] WARNING fablab/kernel.bootstrapBinding: no binding [transitB_host_004_private_ip]
[   0.001] WARNING fablab/kernel.bootstrapBinding: no binding [terminator_host_003_public_ip]
[   0.001] WARNING fablab/kernel.bootstrapBinding: no binding [terminator_host_003_private_ip]
[   0.001]    INFO fablab/kernel/terraform.(*terraformVisitor).visit: => [main.tf]
[   0.002]    INFO fablab/kernel/process.(*Process).Run: executing [terraform init]
Initializing modules...
Downloading /home/michael/Repos/nf/fablab/tf/instance for initiator_host_001...
- initiator_host_001 in .terraform/modules/initiator_host_001

  	< terraform output trimmed >

Outputs:

initiator_host_001_private_ip = 10.0.0.164
initiator_host_001_public_ip = 18.234.38.64
initiator_host_ctrl_private_ip = 10.0.0.196
initiator_host_ctrl_public_ip = 54.152.114.125
terminator_host_003_private_ip = 10.0.0.48
terminator_host_003_public_ip = 34.222.241.64
transitA_host_002_private_ip = 10.0.0.64
transitA_host_002_public_ip = 13.52.75.117
transitB_host_004_private_ip = 10.0.0.182
transitB_host_004_public_ip = 54.203.79.229
[ 315.280]    INFO fablab/kernel/process.(*Process).Run: executing [terraform output initiator_host_ctrl_public_ip]
[ 315.314]    INFO fablab/kernel/terraform.(*terraform).bind: set public ip [54.152.114.125] for [initiator/ctrl]
[ 315.314]    INFO fablab/kernel/process.(*Process).Run: executing [terraform output initiator_host_ctrl_private_ip]
[ 315.346]    INFO fablab/kernel/terraform.(*terraform).bind: set private ip [10.0.0.196] for [initiator/ctrl]
[ 315.346]    INFO fablab/kernel/process.(*Process).Run: executing [terraform output initiator_host_001_public_ip]
[ 315.376]    INFO fablab/kernel/terraform.(*terraform).bind: set public ip [18.234.38.64] for [initiator/001]
[ 315.376]    INFO fablab/kernel/process.(*Process).Run: executing [terraform output initiator_host_001_private_ip]
[ 315.408]    INFO fablab/kernel/terraform.(*terraform).bind: set private ip [10.0.0.164] for [initiator/001]
[ 315.408]    INFO fablab/kernel/process.(*Process).Run: executing [terraform output transitA_host_002_public_ip]
[ 315.440]    INFO fablab/kernel/terraform.(*terraform).bind: set public ip [13.52.75.117] for [transitA/002]
[ 315.441]    INFO fablab/kernel/process.(*Process).Run: executing [terraform output transitA_host_002_private_ip]
[ 315.473]    INFO fablab/kernel/terraform.(*terraform).bind: set private ip [10.0.0.64] for [transitA/002]
[ 315.474]    INFO fablab/kernel/process.(*Process).Run: executing [terraform output transitB_host_004_public_ip]
[ 315.504]    INFO fablab/kernel/terraform.(*terraform).bind: set public ip [54.203.79.229] for [transitB/004]
[ 315.505]    INFO fablab/kernel/process.(*Process).Run: executing [terraform output transitB_host_004_private_ip]
[ 315.536]    INFO fablab/kernel/terraform.(*terraform).bind: set private ip [10.0.0.182] for [transitB/004]
[ 315.536]    INFO fablab/kernel/process.(*Process).Run: executing [terraform output terminator_host_003_public_ip]
[ 315.567]    INFO fablab/kernel/terraform.(*terraform).bind: set public ip [34.222.241.64] for [terminator/003]
[ 315.567]    INFO fablab/kernel/process.(*Process).Run: executing [terraform output terminator_host_003_private_ip]
[ 315.599]    INFO fablab/kernel/terraform.(*terraform).bind: set private ip [10.0.0.48] for [terminator/003]  	
```

`fablab` now wraps `terraform` in these examples, and will automatically do the `fablab init`, the `fablab apply`, and will bind the infrastructure outputs into the model.

## Building the Run Configuration

With the infrastructure expressed, we can now build the configuration for the environment from the bound model.

```
$ fablab build
[   0.000]    INFO fablab/kernel.resolveRunPath: resolved run path [/home/michael/Repos/nf/fablab/build/run]
[   0.000]    INFO fablab/kernel/pki.generateCa: [/home/michael/Repos/nf/ziti/bin/ziti pki create ca --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name root --ca-file root]
[   1.898]    INFO fablab/kernel/pki.generateCa: [/home/michael/Repos/nf/ziti/bin/ziti pki create intermediate --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name root]
[   2.685]    INFO fablab/kernel/pki.(*fabric).Configure: generating public ip identity [003/003] on [terminator/003]
[   2.685]    INFO fablab/kernel/pki.generateCert: generating certificate [003:34.222.241.64]
[   2.685]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create key --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --key-file 003]
[   4.593]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create server --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --server-file 003-server --ip 34.222.241.64 --key-file 003]
[   4.612]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create client --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --client-file 003-client --key-file 003 --client-name 003]
[   4.627]    INFO fablab/kernel/pki.(*fabric).Configure: generating public ip identity [001/001] on [initiator/001]
[   4.627]    INFO fablab/kernel/pki.generateCert: generating certificate [001:18.234.38.64]
[   4.627]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create key --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --key-file 001]
[   5.482]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create server --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --server-file 001-server --ip 18.234.38.64 --key-file 001]
[   5.499]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create client --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --client-file 001-client --key-file 001 --client-name 001]
[   5.516]    INFO fablab/kernel/pki.(*fabric).Configure: generating public ip identity [ctrl/ctrl] on [initiator/ctrl]
[   5.516]    INFO fablab/kernel/pki.generateCert: generating certificate [ctrl:54.152.114.125]
[   5.516]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create key --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --key-file ctrl]
[   7.191]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create server --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --server-file ctrl-server --ip 54.152.114.125 --key-file ctrl]
[   7.208]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create client --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --client-file ctrl-client --key-file ctrl --client-name ctrl]
[   7.232]    INFO fablab/kernel/pki.(*fabric).Configure: generating public ip identity [002/002] on [transitA/002]
[   7.232]    INFO fablab/kernel/pki.generateCert: generating certificate [002:13.52.75.117]
[   7.232]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create key --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --key-file 002]
[   8.017]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create server --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --server-file 002-server --ip 13.52.75.117 --key-file 002]
[   8.033]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create client --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --client-file 002-client --key-file 002 --client-name 002]
[   8.050]    INFO fablab/kernel/pki.(*fabric).Configure: generating public ip identity [004/004] on [transitB/004]
[   8.050]    INFO fablab/kernel/pki.generateCert: generating certificate [004:54.203.79.229]
[   8.050]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create key --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --key-file 004]
[   9.353]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create server --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --server-file 004-server --ip 54.203.79.229 --key-file 004]
[   9.369]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create client --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --client-file 004-client --key-file 004 --client-name 004]
[   9.392]    INFO fablab/kernel/pki.generateCert: generating certificate [dotziti:127.0.0.1]
[   9.392]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create key --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --key-file dotziti]
[  12.421]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create server --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --server-file dotziti-server --ip 127.0.0.1 --key-file dotziti]
[  12.437]    INFO fablab/kernel/pki.generateCert: [/home/michael/Repos/nf/ziti/bin/ziti pki create client --pki-root /home/michael/Repos/nf/fablab/build/run/pki --ca-name intermediate --client-file dotziti-client --key-file dotziti --client-name dotziti]
[  12.454]    INFO fablab/kernel/pki.generateLocalIdentities: config => [local_identities.yml]
[  12.455]    INFO fablab/kernel/config.(*componentConfig).generateConfigForHost: config [ctrl.yml] => [ctrl.yml]
[  12.455]    INFO fablab/kernel/config.(*componentConfig).generateConfigForHost: config [ingress_router.yml] => [001.yml]
[  12.455]    INFO fablab/kernel/config.(*componentConfig).generateConfigForHost: config [transit_router.yml] => [002.yml]
[  12.456]    INFO fablab/kernel/config.(*componentConfig).generateConfigForHost: config [transit_router.yml] => [004.yml]
[  12.456]    INFO fablab/kernel/config.(*componentConfig).generateConfigForHost: config [egress_router.yml] => [003.yml]
[  12.456]    INFO fablab/kernel/config.(*staticConfig).Configure: [/home/michael/Repos/nf/fablab/templates/cfg/10-ambient.loop2.yml] => [/home/michael/Repos/nf/fablab/build/run/cfg/10-ambient.loop2.yml]
[  12.456]    INFO fablab/kernel/config.(*staticConfig).Configure: [/home/michael/Repos/nf/fablab/templates/cfg/remote_identities.yml] => [/home/michael/Repos/nf/fablab/build/run/cfg/remote_identities.yml]
```

The run kit is now fully built, and contains all of the configuration and PKI infrastructure required to implement the model onto the environment.

## Kitting

"Kitting" assembles all of the component parts into a single, distributable tree suitable for synchronization onto the hosts in your environment.

```
$ fablab kit
[   0.000]    INFO fablab/kernel.resolveRunPath: resolved run path [/home/michael/Repos/nf/fablab/build/run]
```

The output of your kitting will end up at `${FABLAB_RUN}/kit`

## Synchronize

Next, we'll synchronize the kit onto all of the hosts in our environment;

```
$ fablab sync
[   0.000]    INFO fablab/kernel.resolveRunPath: resolved run path [/home/michael/Repos/nf/fablab/build/run]
[   0.000]    INFO fablab/kernel/semaphore.(*restartStage).Distribute: waiting for expressed hosts to restart (pre-delay: 1m30s)
[  90.001]    INFO fablab/kernel/semaphore.(*restartStage).Distribute: starting restart checks
[  90.001]    INFO fablab/kernel.RemoteExec: executing [54.152.114.125]: 'uptime'
[  90.721]    INFO fablab/kernel/semaphore.(*restartStage).Distribute: 20:58:38 up 3 min,  0 users,  load average: 0.02, 0.06, 0.02
[  90.721]    INFO fablab/kernel.RemoteExec: executing [18.234.38.64]: 'uptime'
[  91.648]    INFO fablab/kernel/semaphore.(*restartStage).Distribute: 20:58:39 up 3 min,  0 users,  load average: 0.03, 0.06, 0.03
[  91.648]    INFO fablab/kernel.RemoteExec: executing [13.52.75.117]: 'uptime'
[  93.046]    INFO fablab/kernel/semaphore.(*restartStage).Distribute: 20:58:40 up 3 min,  0 users,  load average: 0.17, 0.07, 0.02
[  93.047]    INFO fablab/kernel.RemoteExec: executing [54.203.79.229]: 'uptime'
[  95.344]    INFO fablab/kernel/semaphore.(*restartStage).Distribute: 20:58:43 up 1 min,  0 users,  load average: 0.07, 0.05, 0.02
[  95.344]    INFO fablab/kernel.RemoteExec: executing [34.222.241.64]: 'uptime'
[  97.090]    INFO fablab/kernel/semaphore.(*restartStage).Distribute: 20:58:44 up 3 min,  0 users,  load average: 0.01, 0.02, 0.00
[  97.090]    INFO fablab/kernel.RemoteExec: executing [54.203.79.229]: 'mkdir -p /home/fedora/fablab'
[  98.354]    INFO fablab/kernel/process.(*Process).Run: executing [rsync -avz -e ssh -o "StrictHostKeyChecking no" --delete /home/michael/Repos/nf/fablab/build/run/kit/ fedora@54.203.79.229:/home/fedora/fablab]
Warning: Permanently added '54.203.79.229' (ECDSA) to the list of known hosts.
sending incremental file list
./
bin/
bin/ziti

	< rsync output trimmed >

pki/root/certs/
pki/root/certs/intermediate.cert
pki/root/certs/root.cert
pki/root/keys/
pki/root/keys/intermediate.key
pki/root/keys/root.key

sent 46,611,732 bytes  received 1,268 bytes  593,796.18 bytes/sec
total size is 99,298,180  speedup is 2.13
```

The above output has been shortened a great deal. You'll see lots of output from the synchronization onto your environment.

## Activate

Now we can activate our geo-scale fabric example. We'll use `fablab activate` do this. This will run both the `bootstrap` and `start` actions in sequence.

```
$ fablab activate
[   0.000]    INFO fablab/kernel.resolveRunPath: resolved run path [/home/michael/Repos/nf/fablab/build/run]
[   0.001]    INFO fablab/kernel.RemoteExec: executing [54.152.114.125]: 'ps x'
[   0.553]    INFO fablab/kernel.RemoteExec: executing [54.152.114.125]: 'nohup /home/fedora/fablab/bin/ziti-controller --log-formatter pfxlog run /home/fedora/fablab/cfg/ctrl.yml > ziti-controller.log 2>&1 &'
[   0.988]    INFO fablab/actions.(*sleep).Execute: sleeping for [2s]
[   2.989]    INFO fablab/actions.(*fabricCli).Execute: [/home/michael/Repos/nf/ziti/bin/ziti-fabric create router /home/michael/Repos/nf/fablab/build/run/pki/intermediate/certs/001-client.cert -i fablab]
[   3.149]    INFO fablab/actions.(*fabricCli).Execute: out:[success], err:[]
[   3.150]    INFO fablab/actions.(*fabricCli).Execute: [/home/michael/Repos/nf/ziti/bin/ziti-fabric create router /home/michael/Repos/nf/fablab/build/run/pki/intermediate/certs/002-client.cert -i fablab]
[   3.307]    INFO fablab/actions.(*fabricCli).Execute: out:[success], err:[]
[   3.307]    INFO fablab/actions.(*fabricCli).Execute: [/home/michael/Repos/nf/ziti/bin/ziti-fabric create router /home/michael/Repos/nf/fablab/build/run/pki/intermediate/certs/004-client.cert -i fablab]
[   3.461]    INFO fablab/actions.(*fabricCli).Execute: out:[success], err:[]
[   3.461]    INFO fablab/actions.(*fabricCli).Execute: [/home/michael/Repos/nf/ziti/bin/ziti-fabric create router /home/michael/Repos/nf/fablab/build/run/pki/intermediate/certs/003-client.cert -i fablab]
[   3.627]    INFO fablab/actions.(*fabricCli).Execute: out:[success], err:[]
[   3.627]    INFO fablab/actions.(*fabricCli).Execute: [/home/michael/Repos/nf/ziti/bin/ziti-fabric create service loop tcp:127.0.0.1:8171 003 -i fablab]
[   3.785]    INFO fablab/actions.(*fabricCli).Execute: out:[success], err:[]
[   3.785]    INFO fablab/kernel.RemoteExec: executing [54.152.114.125]: 'mkdir -p /home/fedora/.ziti'
[   4.234]    INFO fablab/kernel.RemoteExec: executing [54.152.114.125]: 'rm -f /home/fedora/.ziti/identities.yml'
[   4.690]    INFO fablab/kernel.RemoteExec: executing [54.152.114.125]: 'ln -s /home/fedora/fablab/cfg/remote_identities.yml /home/fedora/.ziti/identities.yml'
[   5.118]    INFO fablab/kernel.RemoteExec: executing [18.234.38.64]: 'mkdir -p /home/fedora/.ziti'
[   5.650]    INFO fablab/kernel.RemoteExec: executing [18.234.38.64]: 'rm -f /home/fedora/.ziti/identities.yml'
[   6.071]    INFO fablab/kernel.RemoteExec: executing [18.234.38.64]: 'ln -s /home/fedora/fablab/cfg/remote_identities.yml /home/fedora/.ziti/identities.yml'
[   6.476]    INFO fablab/kernel.RemoteExec: executing [13.52.75.117]: 'mkdir -p /home/fedora/.ziti'
[   7.730]    INFO fablab/kernel.RemoteExec: executing [13.52.75.117]: 'rm -f /home/fedora/.ziti/identities.yml'
[   8.900]    INFO fablab/kernel.RemoteExec: executing [13.52.75.117]: 'ln -s /home/fedora/fablab/cfg/remote_identities.yml /home/fedora/.ziti/identities.yml'
[  10.096]    INFO fablab/kernel.RemoteExec: executing [54.203.79.229]: 'mkdir -p /home/fedora/.ziti'
[  11.423]    INFO fablab/kernel.RemoteExec: executing [54.203.79.229]: 'rm -f /home/fedora/.ziti/identities.yml'
[  12.700]    INFO fablab/kernel.RemoteExec: executing [54.203.79.229]: 'ln -s /home/fedora/fablab/cfg/remote_identities.yml /home/fedora/.ziti/identities.yml'
[  13.973]    INFO fablab/kernel.RemoteExec: executing [34.222.241.64]: 'mkdir -p /home/fedora/.ziti'
[  15.387]    INFO fablab/kernel.RemoteExec: executing [34.222.241.64]: 'rm -f /home/fedora/.ziti/identities.yml'
[  16.770]    INFO fablab/kernel.RemoteExec: executing [34.222.241.64]: 'ln -s /home/fedora/fablab/cfg/remote_identities.yml /home/fedora/.ziti/identities.yml'
[  18.155]    INFO fablab/kernel.RemoteExec: executing [54.152.114.125]: 'ps x'
[  18.599]    INFO fablab/kernel.RemoteKill: line [  974 ?        Sl     0:00 /home/fedora/fablab/bin/ziti-controller --log-formatter pfxlog run /home/fedora/fablab/cfg/ctrl.yml]
[  18.599]    INFO fablab/kernel.RemoteExec: executing [54.152.114.125]: 'kill 974'
[  19.048]    INFO fablab/kernel.RemoteExec: executing [54.152.114.125]: 'nohup /home/fedora/fablab/bin/ziti-controller --log-formatter pfxlog run /home/fedora/fablab/cfg/ctrl.yml > ziti-controller.log 2>&1 &'
[  19.500]    INFO fablab/actions.(*sleep).Execute: sleeping for [2s]
[  21.500]    INFO fablab/kernel.RemoteExec: executing [54.203.79.229]: 'nohup /home/fedora/fablab/bin/ziti-router --log-formatter pfxlog run /home/fedora/fablab/cfg/004.yml > ziti-router.log 2>&1 &'
[  22.820]    INFO fablab/kernel.RemoteExec: executing [34.222.241.64]: 'nohup /home/fedora/fablab/bin/ziti-router --log-formatter pfxlog run /home/fedora/fablab/cfg/003.yml > ziti-router.log 2>&1 &'
[  24.200]    INFO fablab/kernel.RemoteExec: executing [18.234.38.64]: 'nohup /home/fedora/fablab/bin/ziti-router --log-formatter pfxlog run /home/fedora/fablab/cfg/001.yml > ziti-router.log 2>&1 &'
[  24.616]    INFO fablab/kernel.RemoteExec: executing [13.52.75.117]: 'nohup /home/fedora/fablab/bin/ziti-router --log-formatter pfxlog run /home/fedora/fablab/cfg/002.yml > ziti-router.log 2>&1 &'
[  25.806]    INFO fablab/actions.(*sleep).Execute: sleeping for [2s]
[  27.806]    INFO fablab/kernel.RemoteExec: executing [34.222.241.64]: 'nohup /home/fedora/fablab/bin/ziti-fabric-test loop2 listener > /home/fedora/ziti-fabric-test.log 2>&1 &'
[  29.209]    INFO fablab/actions.(*sleep).Execute: sleeping for [2s]
[  31.210]    INFO fablab/kernel.RemoteExec: executing [18.234.38.64]: 'nohup /home/fedora/fablab/bin/ziti-fabric-test loop2 dialer /home/fedora/fablab/cfg/10-ambient.loop2.yml -e tls:18.234.38.64:7001 > /home/fedora/ziti-fabric-test.log 2>&1 &'
```

Verify that your new geo-scale network has active sessions on it, using `ziti-fabric` (make sure `ZITI_ROOT/bin` is in your `PATH`):

```
$ ziti-fabric -i fablab list sessions

Sessions: (11)

Id           | Client       | Service      | Path
93ZK         | dotziti      | loop         | [r/001]->{l/OYjW}->[r/002]->{l/X9aX}->[r/003]
9pQv         | dotziti      | loop         | [r/001]->{l/WNKm}->[r/004]->{l/W0NX}->[r/003]
K6eX         | dotziti      | loop         | [r/001]->{l/OYjW}->[r/002]->{l/X9aX}->[r/003]
K76X         | dotziti      | loop         | [r/001]->{l/OYjW}->[r/002]->{l/X9aX}->[r/003]
KOD9         | dotziti      | loop         | [r/001]->{l/WNKm}->[r/004]->{l/W0NX}->[r/003]
Kx59         | dotziti      | loop         | [r/001]->{l/OYjW}->[r/002]->{l/X9aX}->[r/003]
XPaK         | dotziti      | loop         | [r/001]->{l/OYjW}->[r/002]->{l/X9aX}->[r/003]
Xdkv         | dotziti      | loop         | [r/001]->{l/OYjW}->[r/002]->{l/X9aX}->[r/003]
XwBv         | dotziti      | loop         | [r/001]->{l/OYjW}->[r/002]->{l/X9aX}->[r/003]
vD6K         | dotziti      | loop         | [r/001]->{l/OYjW}->[r/002]->{l/X9aX}->[r/003]
vbPK         | dotziti      | loop         | [r/001]->{l/OYjW}->[r/002]->{l/X9aX}->[r/003]

```

### Dispose

Once you've explored your model, and you'd like to dispose of it (releasing all of the associated cloud resources), you'll want to use `fablab dispose`.

```
$ fablab dispose
[   0.000]    INFO fablab/kernel.resolveRunPath: resolved run path [/home/michael/Repos/nf/fablab/build/run]
[   0.000]    INFO fablab/kernel/process.(*Process).Run: executing [terraform destroy -auto-approve]
module.initiator_region.aws_vpc.fablab: Refreshing state... [id=vpc-06936889eb8231a7b]
module.transitA_region.aws_vpc.fablab: Refreshing state... [id=vpc-081fc1936bdfec03c]
module.terminator_region.aws_vpc.fablab: Refreshing state... [id=vpc-0528b2ee52d104840]
module.transitB_region.aws_vpc.fablab: Refreshing state... [id=vpc-0789db81b798fd34e]
module.initiator_region.aws_internet_gateway.fablab: Refreshing state... [id=igw-0124dc5960694666c]
module.initiator_region.aws_subnet.fablab: Refreshing state... [id=subnet-0663bcb17174ea51d]
module.initiator_region.aws_security_group.fablab: Refreshing state... [id=sg-0b0acc0ab9c1ff6ef]
module.initiator_region.aws_route_table.fablab: Refreshing state... [id=rtb-0256b64844d69d1b9]
module.initiator_host_001.aws_instance.fablab: Refreshing state... [id=i-0bb1592fa90dbe851]
module.initiator_region.aws_route_table_association.fablab: Refreshing state... [id=rtbassoc-01b7bfa6c347c1ea2]
module.transitA_region.aws_internet_gateway.fablab: Refreshing state... [id=igw-00cbd6cdbe87c7c5f]
module.transitA_region.aws_subnet.fablab: Refreshing state... [id=subnet-01f0a47fcf02f0458]
module.transitA_region.aws_security_group.fablab: Refreshing state... [id=sg-07b26cdfdcb3ca5dc]
module.transitA_region.aws_route_table.fablab: Refreshing state... [id=rtb-06774b59daf7f10dc]
module.transitA_host_002.aws_instance.fablab: Refreshing state... [id=i-07b3233f45ff4f9e6]
module.transitB_region.aws_internet_gateway.fablab: Refreshing state... [id=igw-05af2d2ea54a246b3]
module.transitB_region.aws_subnet.fablab: Refreshing state... [id=subnet-0e600750777aa4f8f]
module.transitB_region.aws_security_group.fablab: Refreshing state... [id=sg-01aac2952abcbf498]
module.terminator_region.aws_internet_gateway.fablab: Refreshing state... [id=igw-0ee1e9f5db86726a6]

	< terraform destroy output shortened >

module.initiator_region.aws_security_group.fablab: Destroying... [id=sg-0b0acc0ab9c1ff6ef]
module.initiator_region.aws_security_group.fablab: Destruction complete after 1s
module.initiator_region.aws_subnet.fablab: Destruction complete after 1s
module.initiator_region.aws_vpc.fablab: Destroying... [id=vpc-06936889eb8231a7b]
module.initiator_region.aws_vpc.fablab: Destruction complete after 0s
module.transitB_region.aws_internet_gateway.fablab: Destruction complete after 1m13s
module.transitB_host_004.aws_instance.fablab: Destruction complete after 1m13s
module.transitB_region.aws_subnet.fablab: Destroying... [id=subnet-0e600750777aa4f8f]
module.transitB_region.aws_security_group.fablab: Destroying... [id=sg-01aac2952abcbf498]
module.transitB_region.aws_security_group.fablab: Destruction complete after 1s
module.transitB_region.aws_subnet.fablab: Destruction complete after 1s
module.transitB_region.aws_vpc.fablab: Destroying... [id=vpc-0789db81b798fd34e]
module.transitB_region.aws_vpc.fablab: Destruction complete after 0s

Destroy complete! Resources: 29 destroyed.
```