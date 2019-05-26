## AliYun ECS Control

### Install

Clone the repo and add the following alias to ~/.bash_profile
```
alias ecs='bash PATH_TO/ecs/scripts/ecs.sh'
```
Then source ~/.bash_profile

### Usage

```
ecs up     # create a new instance or start an existing one
ecs down   # stop an existing instance
ecs del    # deletee an instance
ecs desc   # only list available instances
```

Instance related configs are in [config.go](https://github.com/iamjinlei/ecs/blob/master/config.go)
