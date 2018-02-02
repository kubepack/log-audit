# Welcome

Log-audit server is a log server for [kubepack](https://github.com/kubepack/kubepack). Log audit uses webhook and audit to communicate with kubernetes apiserver.
kubernetes apiserver sends events to log-audit server using webhook. 
Then, log-audit server saves these events to database(log-audit uses [goleveldb](https://github.com/syndtr/goleveldb) for store data).

### Store Process

