# kubectl-multi-exec

## Installing

Copy the `kubectl-multi_exec` binary to the your `PATH` and ensure it has execution permissions.  
You can use the `kubectl plugin` command to check if the installation was successful.

```
❯ kubectl plugin list
The following compatible plugins are available:

/usr/local/bin/kubectl-multi_exec
```

## Getting Started

Specify the target label in the `selector` option.  
This example of execution a command on a pod which has the label `app=test`.

```shell
❯ kubectl get pod -l 'app=test'
NAME    READY   STATUS    RESTARTS   AGE
test    1/1     Running   0          25s
test2   1/1     Running   0          18s

❯ kubectl multi-exec --selector 'app=test' -- hostname
test
test2

❯ kubectl multi-exec --selector 'app=test' -- uname -a
Linux test 4.15.0-99-generic #100-Ubuntu SMP Wed Apr 22 20:32:56 UTC 2020 x86_64 x86_64 x86_64 GNU/Linux
Linux test2 4.15.0-99-generic #100-Ubuntu SMP Wed Apr 22 20:32:56 UTC 2020 x86_64 x86_64 x86_64 GNU/Linux
```
