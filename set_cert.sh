mkdir -p certificates
sudo cp /etc/letsencrypt/live/team-gg.net/fullchain.pem certificates/cert.pem
sudo cp /etc/letsencrypt/live/team-gg.net/privkey.pem certificates/key.pem
sudo chmod 664 certificates/cert.pem
sudo chmod 664 certificates/key.pem