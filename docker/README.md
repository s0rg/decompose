# decompose

Reverse-engineering tool for docker environments.

# how to run

# scan containers
```
docker run \
        -v /var/run/docker.sock:/var/run/docker.sock \
        -v /:/rootfs:ro \
        -e IN_DOCKER_PROC_ROOT=/rootfs \
        s0rg/decompose:latest > mystream.json
```

# process results
```
docker run \
        s0rg/decompose:latest -load mystream.json -format sdsl > workspace.dsl
```


[more options and documentaion](https://github.com/s0rg/decompose)
