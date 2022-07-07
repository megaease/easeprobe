# EaseProbe Test

## 1 Benchmark Test

### 1.1 Generate the configuration files

```ssh
./gen.sh
```
these files would be generated

- TCP probe configuration file
  - `config.tcp.100.yaml`
  - `config.tcp.1k.yaml`
  - `config.tcp.2k.yaml`
  - `config.tcp.5k.yaml`
  - `config.tcp.10k.yaml`
  - `config.tcp.20k.yaml`

- HTTP probe configuration file
  - `config.http.100.yaml`
  - `config.http.1k.yaml`
  - `config.http.2k.yaml`
  - `config.http.5k.yaml`
  - `config.http.10k.yaml`
  - `config.http.20k.yaml`

### 1.2. Run the EaseProbe

You can edit the config file's to modify probe intervals.

```shell
./easeprobe -d -f config.tcp.10k.yaml > easeprobe.log 2>&1
```

### 1.3. Others

The test data - `data.cvs` is extracted from the top 1M website, the original file can be downloaded via this URL: https://downloads.majestic.com/majestic_million.csv