#!/usr/bin/env bash
cd $(dirname $0)

echo -e "Create volumes: sentry-data, sentry-postgres\n"

volumes=$(docker volume ls)

if [[ $volumes == *"sentry-data"* ]]; then
   echo "sentry-data already exist."
else
    docker volume create --name=sentry-data
fi

if [[ $volumes == *"sentry-postgres"* ]]; then
   echo "sentry-postgres already exist."
else
    docker volume create --name=sentry-postgres
fi

echo -e "\nGenerate private key\n"
if [ ! -f ".env" ]; then
    > .env
fi
if grep -q SENTRY_SECRET_KEY ".env"; then
    echo "SENTRY_SECRET_KEY already setup"
else
    pKey=$(docker-compose run --rm web config generate-secret-key)
    echo "SENTRY_SECRET_KEY=$pKey" > .env
fi

echo -e "\nBuild Sentry database\n"
docker-compose run --rm web upgrade

echo -e "\nGet network address\n"
ip=$(docker network inspect lachesis | grep Gateway)
ip=${ip//\"/}
ip=${ip//"Gateway: "/}
echo $ip

echo -e "\nAssembly client DSN\n"
container_id=$(docker ps -aqf "name=sentry_postgres_1")
row=$(docker exec $container_id psql -U postgres -d postgres -c "SELECT public_key FROM sentry_projectkey WHERE project_id=1;")
pubKey=$(echo $row | awk '{print $3}')
dsn="http://$pubKey@$ip:9000/1"
echo "dsn=$dsn" > .dsn

echo -e "\nStart Setry\n"
docker-compose up -d
