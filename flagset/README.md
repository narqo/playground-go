# flag.FlagSet

See [flag][].

Get help:

```
> go run ./ -help
Usage: flagset [options] <commands> <args>

Options
  -namespace string
    	Set namespace.
  -version
    	Print version and exit.
  -work-tree string
    	Set the path to the working tree.

Commands
  add
	Add command.
  archive
	Archive command.
exit status 2
```

Get a command help:

```
> go run ./ add -help
...
```

Run a command:

```
>  go run ./ -namespace ns1 add -verbose -interactive /tmp/1.txt
...
```

[flag]: https://godoc.org/flag
