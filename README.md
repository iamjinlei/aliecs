## AliYun ECS Control

This is a utilitiy tool allows you to up/down AliYun ecs instance conveniently, saving instance cost by bringing it up on demand. 

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
ecs del    # delete an instance
ecs desc   # list available instances
ecs go     # ssh into one of the instances
```

Instance related configs are in [config.go](https://github.com/iamjinlei/ecs/blob/master/config.go)
