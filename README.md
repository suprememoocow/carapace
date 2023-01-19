# Carapace

Carapace is a small tool for working with shelly devices on your local network. 

## Query

`carapace query` queries across all devices, using a JQ-like expression to format the output.

```console
$ carapace query '{mac: .settings.device.mac, name:.settings.name}'
{"mac":"XXXXXXX39FC4","name":"Awning Light String"}
{"mac":"XXXXXXX80A91","name":"Borehole Pump"}
{"mac":"XXXXXXX8B482","name":"Piano Hallway"}
...
```