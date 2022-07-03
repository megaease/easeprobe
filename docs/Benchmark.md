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

The document is the performance test report for EaseProbe.

## 1. Test Environment

- OS: Ubuntu 20.04 LTS
- VM: AWS EC2 C5.2xlarge -  8 vCPUs, 16 GiB RAM, EBS Volume - 100 GiB
- Config: Everything is default configured. only enlarge the max open file limit.
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

- Build: the following PRs are important for this test:
  -  [PR #157](https://github.com/megaease/easeprobe/pull/157) - Equally distributing start the probers.
     - without this PR, each prober would start to work a same time.
     - with it the prober would start in evenly distribute with in 1 minute
  -  [PR #159](https://github.com/megaease/easeprobe/pull/159) - Set the `SO_LINGER=0` to tcp, http, tls, ssh probe
     -  without this PR, EaseProbe leave a TIME_WAIT tcp connection.
     -  with this PR, EaseProbe will close the TIME_WAIT tcp connection.


## 2. Test Cases

Only ONE Single EaseProbe Process for this performance test.

- Only test HTTPS and TCP probe
- with 100, 1000, 2000, 5000 & 10,0000 probers
- with 5s, 10s, 30s, 60s probe interval
- With 5 mins running, and check the CPU, Memory, and Network connection.

## 3. Test Data

(tbd)

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
| https | 5s       |   18.7%    |   105MB           |    50       |
| https | 10s      |   10.1%    |    98MB           |    24       |
| https | 30s      |    4.0%    |    75MB           |     8       |
| https | 60s      |    2.0%    |    75MB           |     4       |

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| tcp   | 5s       |    3.3%    |   105MB           |    48       |
| tcp   | 10s      |    1.3%    |   100MB           |    24       |
| tcp   | 30s      |    0.7%    |    80MB           |     8       |
| tcp   | 60s      |    0.7%    |    79MB           |     4       |


### 4.2  With 2,000 Probers

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| https | 5s       |   38.5%    |   183MB           |    72       |
| https | 10s      |   20.2%    |   150MB           |    36       |
| https | 30s      |    8.0%    |   125MB           |    12       |
| https | 60s      |    5.3%    |   114MB           |     6       |

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| tcp   | 5s       |    5.3%    |   165MB           |    72       |
| tcp   | 10s      |    2.7%    |   150MB           |    36       |
| tcp   | 30s      |    1.3%    |   145MB           |    12       |
| tcp   | 60s      |    0.7%    |   142MB           |     6       |

### 4.3  With 5,000 Probers

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| https | 5s       |   88.4%    |   580MB           |   227       |
| https | 10s      |   43.6%    |   380MB           |   116       |
| https | 30s      |   18.7%    |   270MB           |    56       |
| https | 60s      |   10.0%    |   254MB           |    18       |

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| tcp   | 5s       |   12.3%    |   380MB           |   198       |
| tcp   | 10s      |    8.7%    |   380MB           |   100       |
| tcp   | 30s      |    2.6%    |   370MB           |    32       |
| tcp   | 60s      |    2.3%    |   350MB           |    16       |

### 4.4  With 10,000 Probers

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| https | 5s       |  100.0% *  |   2.5GB           |   241       |
| https | 10s      |   88.6%    |   960MB           |   136       |
| https | 30s      |   30.7%    |   570MB           |    48       |
| https | 60s      |   17.8%    |   450MB           |    26       |

> Note: https with 5s interval reaches the CPU limitation for all 8 CPU cores, the memory usage is abnormal as well, EaseProbe does not work properly (there are thousands of `SYNC_SENT` and `CLOSE_WAIT `).

| Kind  | Interval | CPU Usage | Memory Usage (RES) | # TIME_WAIT |
| ----- | :------: | :--------: | :---------------: | :---------: |
| tcp   | 5s       |   20.7%    |   800MB           |   222       |
| tcp   | 10s      |   10.3%    |   760MB           |   110       |
| tcp   | 30s      |    3.3%    |   717MB           |    36       |
| tcp   | 60s      |    2.7%    |   492MB           |    18       |