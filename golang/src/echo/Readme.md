
NODE 1 run:
```
./echo -secio -l 10000

```

NODE 2 run: [ replace hash ]
```
echo ./echo -a /ip4/127.0.0.1/tcp/ -l 10001 -d /ip4/10.1.1.2/tcp/10000/ipfs/QmXSkquSVwxFvasF1zVNbNotuVSDRk9QfDyeMrFzZb6oVg -secio
```