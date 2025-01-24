# minitaskx-example
minitaskx Use cases

# Usage
1. start minitaskx depend
```
make init
```

2. start minitaskx server
```
# run worker to execute task
make worker

# run scheduler to schedule task
make scheduler
```

3. use with command line tools
```
make ctl

./minictl -h
```
your can use minictl to:
   - create tasks
   - pause task
   - resume task
   - stop task

4. clean up the resources
make clean

