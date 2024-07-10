# imply (WIP)

A WIP Go tool to generate struct implementations from interfaces.

## installation

``` sh
go install github.com/gleek/imply@latest
```


## usage

```sh
imply <file_name> [interface_name] [struct_name] [package_name]
```

`interface_name` `struct_name` and `package_name` are optional. If not provided, the tool will try to infer them from the file name.


## alternatives
- [impl](https://github.com/josharian/impl)
- [goimpl](https://github.com/sasha-s/goimpl)

## why another tool?
Both impl and goimpl are slow for me. I haven't looked into what would caused that slowness, but figured it would be easy enough to do this myself.
Specially, if I give it the exact place to find the interface.

Also my primary goal was to use this from within Emacs, so that I can quickly generate the struct and move it to the correct place.

**Note**: this is an extremely rudimentary and experimental implementation and I'm not sure if it work for all the cases. Use with caution.
