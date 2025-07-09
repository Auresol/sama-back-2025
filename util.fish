# grap the id of backend-logdy, then logs it
docker logs (docker ps | grep 'backend-logdy' | grep -Po '^[^ ]*')