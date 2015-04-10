# Noop Mode

When in noop mode target configuration files will not be modified.

## Usage

### commandline flag

```
confd -noop
```

### configuration file

```
noop = true
```

### Example

```
confd -onetime -noop
```

-

```
2014-07-08T22:30:10-07:00 confd[16397]: INFO /tmp/myconfig.conf has md5sum c1924fc5c5f2698e2019080b7c043b7a should be 8e76340b541b8ee29023c001a5e4da18
2014-07-08T22:30:10-07:00 confd[16397]: WARNING Noop mode enabled /tmp/myconfig.conf will not be modified
```
