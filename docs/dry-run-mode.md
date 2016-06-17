# Dry-Run Mode

When in dry-run mode target configuration files will not be modified but `check_cmd` will be executed and execution returns error if the execution of `check_cmd` fails.

This mode behaves like `noop` mode but adds the execution of the `check_cmd` command. 

Note: dry-run mode *does not* update target and *does not* run `reload_cmd`.

## Usage

### commandline flag

```
confd -dry-run
```

### configuration file

```
dry-run = true
```

### Example

```
confd -onetime -dry-run
```

-

```
2016-06-09T15:22:28-03:00 confd[29789]: INFO /tmp/myconfig.conf has md5sum 5fc1bd8022b5cbb5aca50b74817fa9c9 should be 7cfb29f5029fc9502f58665d66ce1c6c
2016-06-09T15:22:28-03:00 confd[29789]: WARNING Dry-run mode enabled. /tmp/myconfig.conf will not be modified
```
