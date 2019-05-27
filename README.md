## AliYun ECS Control

This is a utilitiy tool allows you to up/down AliYun ecs instance conveniently, saving instance cost by bringing it up on demand. 

### Install

Clone the repo and add the following alias to ~/.bash_profile
```
alias ecs='bash PATH_TO/ecs/scripts/ecs.sh'
```
Then source ~/.bash_profile

### Usage

Set env vars:
```
export ECS_ACCESS_KEY_ID        # AliYun access key ID
export ECS_ACCESS_KEY_SECRET    # AliYun access key secret
export ECS_KEY_PAIR_NAME        # Optional
export ECS_ROOT_PWD             # Root password
```

Commands:
```
ecs up     # create a new instance or start an existing one
ecs down   # stop an existing instance
ecs del    # delete an instance
ecs desc   # list available instances
ecs go     # ssh into one of the instances
```

Instance related configs are in [config.go](https://github.com/iamjinlei/ecs/blob/master/config.go)
