## vcn untrust

Untrust a digital asset

### Synopsis

Untrust a digital asset

```
vcn untrust [flags]
```

### Options

```
  -a, --attr list     add user defined attributes (format: --attr key=value)
      --hash string   specify the hash of an asset signed by you to untrust, if set no arg(s) can be used
  -h, --help          help for untrust
  -k, --key string    specify which user's key to use for signing, if not set the last available is used
  -p, --public        when signed as public, the asset name and the signer's identity will be visible to everyone
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.vcn/config.json)
```

### SEE ALSO

* [vcn](vcn.md)	 - vChain CodeNotary - code signing in 1 simple step

###### Auto generated by spf13/cobra on 16-May-2019
