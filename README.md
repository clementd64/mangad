# mangad

```yaml
services:
  tachidesk:
    image: ghcr.io/suwayomi/tachidesk
    restart: unless-stopped

  mangad:
    image: ghcr.io/clementd64/mangad
    user: 1000:1000
    command: -url http://tachidesk:4567/ -config /config.yaml -interval 6h -wait-for-it
    working_dir: /manga
    volumes:
      - ./manga:/manga
      - ./config.yaml:/config.yaml
```