# dns-resolver-go

## Usage

A dnslink is a path link in a dns TXT record, like this:
_nkn.foo.com TXT nkn=123abc
For example:
> $ dig TXT _nkn.foo.com \
> _nkn.foo.com.  120   IN  TXT  nkn=123abc

* Add Resolver, DNS link to NKN address
```
account, err := NewAccount(nil)
dnsResolver, err := dnsresolver.NewResolver(nil)
if err != nil {
    return err
}

conf := &nkn.ClientConfig{
    Resolvers: nkngomobile.NewResolverArray(dnsResolver),
}
client, err := NewMultiClient(account, "identifier", 3, true, conf)
client.Send(nkn.NewStringArray("DNS:foo.com"), "Hello world.", nil)
```

