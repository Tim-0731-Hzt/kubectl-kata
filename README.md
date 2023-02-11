# kata kubectl

A `kubectl` plugin to deploy and manage your kata-containers

## Quick Start

### krew installation
```console
foo@bar:~$ kubectl krew install kata
foo@bar:~$ kubectl kata --help
```
### manual installation
```console
foo@bar:~$ make bin
foo@bar:~$ sudo cp ./bin/kata /usr/local/bin
```

## Usage

### Kata deploy

```console
foo@bar:~$ kubectl kata deploy

INFO[0000] create kata-rbac                             
INFO[0000] create kata-deploy                           
INFO[0003] create kata-runtimeclass
INFO[0003] ready to go now
```

### Config debug console

```console
foo@bar:~$ kubectl kata config --debug_console=true
```

### Enter the vm

```console
foo@bar:~$ kubectl kata exec -p POD_NAME -n NAMESPACE

exec called
INFO[0000] validate pod                                 
INFO[0000] Run
                                     
root@clr-5c727e5476d44f809dc5680f8fc06ebb / # 
root@clr-5c727e5476d44f809dc5680f8fc06ebb / # ls
autofs  boot  etc   lib    lost+found  mnt   root  sbin  sys  usr
bin     dev   home  lib64  media       proc  run   srv   tmp  var
```

### Clean up kata

```console
foo@bar:~$ kubectl kata delete

INFO[0000] delete kata-deploy                           
INFO[0031] create kubelet-kata-cleanup                  
INFO[0034] exec cleanup                                 
node/ap-southeast-1.10.0.0.86 unlabeled
INFO[0035] delete kubelet-kata-cleanup                  
INFO[0078] delete kata-rbac                             
INFO[0078] delete kata-runtimeclass
```

## TODO