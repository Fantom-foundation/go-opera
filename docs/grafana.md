# MONITORING OPERA WITH INFLUXDB AND GRAFANA

This tutorial will help you set up monitoring for your Opera node so you can better understand its performance and identify potential problems.

## PREREQUISITES

- You should already be running an instance of Opera.
- Most of the steps and examples are for linux environment, basic terminal knowledge will be helpful

## MONITORING STACK

An Opera client collects lots of data which can be read in the form of a chronological database. To make monitoring easier, you can feed this into data visualisation software. There are multiple options available:

- [Prometheus](https://prometheus.io/) (pull model)
- [InfluxDB](https://www.influxdata.com/get-influxdb/) (push model)
- [Telegraf](https://www.influxdata.com/get-influxdb/)
- [Grafana](https://www.grafana.com/)
- [Datadog](https://www.datadoghq.com/)
- [Chronograf](https://www.influxdata.com/time-series-platform/chronograf/)

In this tutorial, we'll set up your Opera client to push data to **InfluxDB** to create a database and **Grafana** to create a graph visualisation of the data. Doing it manually will help you understand the process better, alter it, and deploy in different environments.

## SETTING UP INFLUXDB

First, let's download and install InfluxDB. Various download options can be found at [Influxdata release page](https://portal.influxdata.com/downloads/). Pick the one that suits your environment. You can also install it from a repository. For example in Debian 10.4 based distribution:

```sh
$ sudo apt update -y && sudo apt upgrade -y
$ sudo apt install curl
$ curl -s https://repos.influxdata.com/influxdb.key | sudo apt-key add -
$ source /etc/os-release
$ echo "deb https://repos.influxdata.com/debian $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/influxdb.list
$ sudo apt-get install influxdb -y
$ sudo systemctl unmask influxdb.service
$ sudo systemctl enable influxdb
$ sudo systemctl start influxdb
```

After successfully installing InfluxDB, make sure it's running on background. By default, it is reachable at `localhost:8086`. Before using influx client, you have to create new user with admin privileges. This user will serve for high level management, creating databases and users.

```
curl -XPOST "http://localhost:8086/query" --data-urlencode "q=CREATE USER fantom WITH PASSWORD 'fantom' WITH ALL PRIVILEGES"
```

Now you can use influx client to enter [InfluxDB shell](https://docs.influxdata.com/influxdb/v1.8/tools/shell/) with this user.


```sh
$ influx -username 'fantom' -password 'fantom'
```

Directly communicating with InfluxDB in its shell, you can create database and user for opera metrics.

```
create database opera
create user opera with password choosepassword
```

Verify created entries with:

```
show databases
show users
```

You cam leave InfluxDB shell by input `exit` into the command.

InfluxDB is running and configured to store metrics from Opera.

## PREPARING OPERA

After setting up database, we need to enable metrics collection in Opera. Pay attention to METRICS AND STATS OPTIONS in `opera --help`. Multiple options can be found there, in this case we want Opera to push data into InfluxDB. Basic setup specifies endpoint where InfluxDB is reachable and authentication for the database.

```
nohup ./build/opera --genesis testnet.g --gcmode=full --metrics --metrics.influxdb --metrics.influxdb.database "opera" --metrics.influxdb.username "opera" --metrics.influxdb.password "chosenpassword" &
```

You can verify that Opera is successfully pushing data, for instance by listing metrics in database. In InfluxDB shell:

```sh
use opera
show measurements
```

## SETTING UP GRAFANA

Next step is installing Grafana which will interpret data graphically. Follow installation process for your environment in Grafana documentation. Make sure to install OSS version if you don't want otherwise. Example installation steps for Debian 10.4 distributions using repository:

```sh
$ curl -tlsv1.3 --proto =https -sL https://packages.grafana.com/gpg.key | sudo apt-key add -
$ echo "deb https://packages.grafana.com/oss/deb stable main" | sudo tee -a /etc/apt/sources.list.d/grafana.list
$ sudo apt update
$ sudo apt install grafana
$ sudo systemctl enable grafana-server
$ udo systemctl start grafana-server
```

When you've got Grafana running, it should be reachable at `localhost:3000`. Use your preferred browser to access this path, then login with the default credentials (user: admin and password: admin). When prompted, change the default password and save.

![Grafana Login Screen](images/grafana_login.png?raw=true "Grafana Login Screen")

You will be redirected to the Grafana home page. First, set up your source data. Click on the configuration icon in the left bar and select "Data sources".

![Grafana Data Source](images/grafana_datasource.png?raw=true "Grafana Data Source")

If there aren't any data sources created yet, click on **[Add data source]** to define one.

![Grafana Create Data Source](images/grafana_create_datasource.png?raw=true "Grafana Create Data Source")

For this setup, select "InfluxDB" and proceed.

![Grafana InfluxDB](images/grafana_influxdb.png?raw=true "Grafana InfluxDB")

Data source configuration is pretty straight forward if you are running tools on the same machine. You need to set the InfluxDB address and details for accessing the database. Refer to the picture below.

![Grafana InfluxDB Settings](images/grafana_influxdb_settings.png?raw=true "Grafana InfluxDB Settings")

If everything is complete and InfluxDB is reachable, click on **[Save and test]** and wait for the confirmation to pop up.

![Grafana InfluxDB Data Source OK](images/grafana_influxdb_ok.png?raw=true "Grafana InfluxDB Data Source OK")

Grafana is now set up to read data from InfluxDB. Now you need to create a dashboard which will interpret and display it. Dashboards properties are encoded in JSON files which can be created by anybody and easily imported. On the left bar, click on **[Import]**.

![Grafana Import](images/grafana_import.png?raw=true "Grafana Import")

For a Opera monitoring dashboard, copy the ID of [this dashboard](https://grafana.com/grafana/dashboards/13877) and paste it in the "Import page" in Grafana. After saving the dashboard, it should look like this:

![Grafana Dashboard](images/grafana_dashboard.png?raw=true "Grafana Dashboard")

We can modify your dashboards. Each panel can be edited, moved, removed or added. You can change your configurations. It's up to you! To learn more about how dashboards work, refer to [Grafana's documentation](https://grafana.com/docs/grafana/latest/dashboards/). You might also be interested in [Alerting](https://grafana.com/docs/grafana/latest/alerting/). This lets you set up alert notifications for when metrics reach certain values. Various communication channels are supported.