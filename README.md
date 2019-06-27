## AliYun ECS Control

This is a utilitiy tool allows you to up/down AliYun ecs instance conveniently, saving instance cost by bringing it up on demand. 

### Install

Clone the repo and add the following alias to ~/.bash_profile
```bash
alias ecs='bash PATH_TO/aliecs/scripts/ecs.sh'
```
Then source ~/.bash_profile

### Usage

Set env vars:
```bash
export ECS_ACCESS_KEY_ID        # AliYun access key ID
export ECS_ACCESS_KEY_SECRET    # AliYun access key secret
export ECS_KEY_PAIR_NAME        # Optional
export ECS_ROOT_PWD             # Root password
```

Commands:
```bash
ecs up     # create a new instance or start an existing one
ecs down   # stop an existing instance
ecs del    # delete an instance
ecs desc   # list available instances
ecs go     # ssh into one of the instances
```
All those commands support an optional index to specify a particular instance to operate on. The index is defined in the table from the **ecs desc**. Index 0 is used by default.

Instance related configs are in [config.go](https://github.com/iamjinlei/aliecs/blob/master/config.go)
