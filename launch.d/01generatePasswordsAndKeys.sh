#!/bin/bash
ENV_DEFAULT=.env.exemple
ENV_OUTPUT=srcs/.env

generate_password() {
    export LC_ALL=C
    echo $(LC_CTYPE=C; tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 32)
}

generate_secret_key() {
    python3 - << EOF
import secrets
length = 50
chars = 'abcdefghijklmnopqrstuvwxyz0123456789%^&*(-_=+)'
secret_key = ''.join(secrets.choice(chars) for i in range(length))
print(secret_key)
EOF
}

> "$ENV_OUTPUT"

while IFS= read -r line; do
    if [[ $line == POSTGRES_PASSWORD* ]]; then
        variable=$(echo "$line" | cut -d '=' -f 1)
        POSTGRES_PASSWORD=$(generate_password)
        echo "$variable=\"$POSTGRES_PASSWORD\"" >> "$ENV_OUTPUT"
    elif [[ $line == DATABASE_URL* ]]; then
        new_url=$(echo "$line" | sed "s/changeme/${POSTGRES_PASSWORD}/")
        echo "$new_url" >> "$ENV_OUTPUT"
    elif [[ $line == "JWT_SECRET"* ]]; then
        variable=$(echo "$line" | cut -d '=' -f 1)
        key=$(generate_secret_key)
        echo "$variable=\"$key\"" >> "$ENV_OUTPUT"
    else
        echo "$line" >> "$ENV_OUTPUT"
    fi
done < "$ENV_DEFAULT"

exit 0
