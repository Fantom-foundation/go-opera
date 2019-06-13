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
if [ -f .env ] && grep -q SENTRY_SECRET_KEY ".env"; then
    echo "SENTRY_SECRET_KEY already setup"
else
    > .env
    pKey=$(docker-compose run --rm web config generate-secret-key)
    echo "SENTRY_SECRET_KEY=$pKey" > .env
fi


echo -e "\nBuild Sentry database\n"
docker-compose run --rm web upgrade


echo -e "\nGet network address\n"
ip=$(docker network inspect -f '{{range .IPAM.Config}}{{.Gateway}}{{end}}' lachesis)
echo $ip


echo -e "\nAssembly client DSN\n"
docker exec sentry_postgres_1 psql \
    -U postgres -d postgres -P "tuples_only" \
    -c "SELECT 'SENTRY_DSN=http://' || public_key || '@${ip}:9000/1' FROM sentry_projectkey WHERE project_id=1;" \
    > .dsn


echo -e "\nStart Setry\n"
docker-compose up -d
