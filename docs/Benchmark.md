# Benchmark

- [Benchmark](#benchmark)
  - [1. Test Environment](#1-test-environment)
  - [2. Test Cases](#2-test-cases)
  - [3. Test Data](#3-test-data)
  - [4. Test Result](#4-test-result)
    - [4.1 With 100 Probers](#41-with-100-probers)
    - [4.2  With 1,000 Probers](#42--with-1000-probers)
    - [4.2  With 2,000 Probers](#42--with-2000-probers)
    - [4.3  With 5,000 Probers](#43--with-5000-probers)
    - [4.4  With 10,000 Probers](#44--with-10000-probers)
    - [4.5 With 20,000 Probers](#45-with-20000-probers)

The document is the performance test report for EaseProbe.

## 1. Test Environment

- **OS**: Ubuntu 20.04 LTS
- **VM**: AWS EC2 C5.2xlarge -  8 vCPUs, 16 GiB RAM, EBS Volume - 100 GiB
- **Config**: Everything is default configured. only enlarge the max open file limit.
  - in `/etc/sysctl.conf`, add

      ```
      fs.file-max = 1000000
      ```
  - in `/etc/security/limits.conf`, add

      ```
      * hard nofile 1000000
      * soft nofile 1000000

      session required pam_limits.so
      ```
  - DNS server is `8.8.8.8`

    > Note: the DNS server `1.1.1.1` would be a problem if we query DNS too many, `1.1.1.1` would block the DNS query.

- **Build**: the following PRs are needed for this test and we must ensure they are included (already merged in `main`):
  -  [PR #157](https://github.com/megaease/easeprobe/pull/157) - Equally distributing start the probers.
     - without this PR, each prober would start to work a same time.
     - with it the prober would start in evenly distribute with in 1 minute
  -  [PR #159](https://github.com/megaease/easeprobe/pull/159) - Set the `SO_LINGER=0` to tcp, http, tls, ssh probe
     -  without this PR, EaseProbe leave a TIME_WAIT tcp connection.
     -  with this PR, EaseProbe will close the TIME_WAIT tcp connection.
  -  [PR #162](https://github.com/megaease/easeprobe/pull/162) - Fix the concurrent map access crashing problem
     -  without this PR, EaseProbe will crash during the test.
     -  with this PR, EaseProbe will not crash during the test.


## 2. Test Cases

Only ONE Single EaseProbe Process for this performance test.

- Only test HTTPS and TCP probe
- with 100, 1000, 2000, 5000 & 10,0000 probers
- with 5s, 10s, 30s, 60s probe interval
- With 5 mins running, and check the CPU, Memory, and Network connection.

Using the the following command to observe the performance:
- `top` and `htop` to check the CPU and Memory usage.
- `netstat  -nat  |  awk  '{print  $6}'  |  sort  |  uniq  -c  |  sort  -n` to check the Network connection.


## 3. Test Data

The test data is taken from the top 1M websites (https://majestic.com/reports/majestic-million)

The test data can be found under the [EaseProbe Test](../resources/test/) folder.

## 4. Test Result

### 4.1 With 100 Probers

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| https | 5s       |    2.0%    |    50MB           |     0       |
| https | 10s      |    1.3%    |    50MB           |     0       |
| https | 30s      |    0.7%    |    45MB           |     0       |
| https | 60s      |    0.3%    |    45MB           |     0       |

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| tcp   | 5s       |    0.7%    |    30MB           |     0       |
| tcp   | 10s      |    0.7%    |    30MB           |     0       |
| tcp   | 30s      |    0.2%    |    30MB           |     0       |
| tcp   | 60s      |    0.2%    |    30MB           |     0       |


### 4.2  With 1,000 Probers

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| https | 5s       |   18.7%    |   105MB           |     0       |
| https | 10s      |   10.1%    |    98MB           |     0       |
| https | 30s      |    4.0%    |    75MB           |     0       |
| https | 60s      |    2.0%    |    75MB           |     0       |

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| tcp   | 5s       |    3.3%    |   105MB           |     0       |
| tcp   | 10s      |    1.3%    |   100MB           |     0       |
| tcp   | 30s      |    0.7%    |    80MB           |     0       |
| tcp   | 60s      |    0.7%    |    79MB           |     0       |


### 4.2  With 2,000 Probers

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| https | 5s       |   38.5%    |   183MB           |     0       |
| https | 10s      |   20.2%    |   150MB           |     0       |
| https | 30s      |    8.0%    |   125MB           |     0       |
| https | 60s      |    5.3%    |   114MB           |     0       |

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| tcp   | 5s       |    5.3%    |   165MB           |     0       |
| tcp   | 10s      |    2.7%    |   150MB           |     0       |
| tcp   | 30s      |    1.3%    |   145MB           |     0       |
| tcp   | 60s      |    0.7%    |   142MB           |     0       |

### 4.3  With 5,000 Probers

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| https | 5s       |   88.4%    |   580MB           |    0       |
| https | 10s      |   43.6%    |   380MB           |    0       |
| https | 30s      |   18.7%    |   270MB           |    0       |
| https | 60s      |   10.0%    |   254MB           |    0       |

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| tcp   | 5s       |   12.3%    |   380MB           |    0       |
| tcp   | 10s      |    8.7%    |   380MB           |    0       |
| tcp   | 30s      |    2.6%    |   370MB           |    0       |
| tcp   | 60s      |    2.3%    |   350MB           |    0       |

### 4.4  With 10,000 Probers

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| https | 5s       |  100.0% *  |   2.5GB *         |    0       |
| https | 10s      |   88.6%    |   960MB           |    0       |
| https | 30s      |   30.7%    |   570MB           |    0       |
| https | 60s      |   17.8%    |   450MB           |    0       |

> Note: https with 5s interval reaches the CPU limitation for all 8 CPU cores, the memory usage is abnormal as well, EaseProbe does not work properly (there are thousands of `SYNC_SENT` and `CLOSE_WAIT `).

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| tcp   | 5s       |   20.7%    |   800MB           |    0       |
| tcp   | 10s      |   10.3%    |   760MB           |    0       |
| tcp   | 30s      |    3.3%    |   717MB           |    0       |
| tcp   | 60s      |    2.7%    |   492MB           |    0       |

### 4.5 With 20,000 Probers

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| https | 5s       |    -       |      -            |   -         |
| https | 10s      |    -       |      -            |   -         |
| https | 20s      |   88.8%    |   1.6GB           |    0        |
| https | 30s      |   58.7%    |   1.3GB           |    0        |
| https | 60s      |   32.6%    |   1.0GB           |    0        |

> Note:  https cannot make the 5s and 10s interval.

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| tcp   | 5s       |   49.7%    |   1.5GB           |   0       |
| tcp   | 10s      |   18.3%    |   1.5GB           |   0       |
| tcp   | 30s      |    4.7%    |   1.5GB           |   0       |
| tcp   | 60s      |    3.4%    |   1.5GB           |   0       |
