go build ./main.go
sudo supervisorctl restart team.gg-server
tail -f output.log