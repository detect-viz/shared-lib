# oracle 腳本測試執行指令
/usr/bin/su - oraedv -c /home/colt_dev/ipoc_conn.sh

# FTP
ftp.alphanetworks.com
basis
cjo40

## grafana deb
sudo apt-get install -y adduser libfontconfig1 musl
wget https://dl.grafana.com/oss/release/grafana_10.2.0_amd64.deb
sudo dpkg -i grafana_10.2.0_amd64.deb

## grafana datatable plugin
docker cp grafana/briangann-datatable-panel/module.js grafana:var/lib/grafana/plugins/briangann-datatable-panel/module.js
docker cp grafana/briangann-datatable-panel/module.js.LICENSE.txt grafana:var/lib/grafana/plugins/briangann-datatable-panel/module.js.LICENSE.txt
docker cp grafana/briangann-datatable-panel/module.js.map grafana:var/lib/grafana/plugins/briangann-datatable-panel/module.js.map

wget -q https://repos.influxdata.com/influxdata-archive_compat.key
echo '393e8779c89ac8d958f81f942f9ad7fb82a25e133faddaf92e15b16e6ac9ce4c influxdata-archive_compat.key' | sha256sum -c && cat influxdata-archive_compat.key | gpg --dearmor | sudo tee /etc/apt/trusted.gpg.d/influxdata-archive_compat.gpg > /dev/null
echo 'deb [signed-by=/etc/apt/trusted.gpg.d/influxdata-archive_compat.gpg] https://repos.influxdata.com/debian stable main' | sudo tee /etc/apt/sources.list.d/influxdata.list
 
sudo apt-get update && sudo apt-get install influxdb2

# 服務
cd /etc/systemd/system
systemctl restart ipoc.service
systemctl status ipoc.service

# mysql 連線
mysql -u bimap -p1qaz2wsx
show databases;

# 查看 firewall 設定
firewall-cmd --list-all
firewall-cmd --permanent --add-port 8086/tcp
firewall-cmd --reload

# 安裝 deb
sudo dpkg -i keycloak.deb
sudo dpkg -i bimap-ipoc.deb
sudo dpkg -i bimap-ipoc-frontend.deb
systemctl daemon-reload

mv /home/parser/upload/alpha/* /home/parser/upload/master
cp /home/parser/upload/master/* /home/bimap/log

# 卸除
sudo dpkg --purge bimap-ipoc-frontend.deb
# 卸載 grafana 及其依賴項：
sudo apt-get remove --auto-remove grafana-enterprise 
# 重置 grafana 密碼
sudo grafana-cli --homepath “/usr/share/grafana/” admin reset-admin-password newpass

# 刪除空檔案
find . -name "*" -type f -size 0c | xargs -n 1 rm -f





sudo apt purge influxdb2
sudo apt purge influxdb2-cli
sudo apt autoclean && sudo apt autoremove
 

rm /etc/apt/sources.list.d/influxdata.list
rm /etc/apt/trusted.gpg.d/influxdata-archive_compat.gpg
sudo rm -rf /etc/influxdb
sudo rm -rf ~/.influxdbv2/configs
sudo rm -rf /var/lib/influxdb
 
sudo apt-get update && sudo apt-get install influxdb2
systemctl restart influxdb

influx s