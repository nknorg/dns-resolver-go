# dns-resolver-go

## Usage

Address in a dns TXT record, like this:

> _nkn.foo.com TXT nkn=123abc

For example:

> $ dig TXT _nkn.foo.com \
> _nkn.foo.com.  120   IN  TXT  nkn=123abc

* Add Resolver, DNS record to NKN address
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

## Use custom username

Address in a dns TXT record use custom username, like this:

> _nkn.user1.foo.com TXT user1@nkn=abc...def \
> _nkn.user2.foo.com TXT user2@nkn=abc...def

```
client.Send(nkn.NewStringArray("DNS:user1@foo.com"), "Hello world.", nil)
client.Send(nkn.NewStringArray("DNS:user2@foo.com"), "Hello world.", nil)
```