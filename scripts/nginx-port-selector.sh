#!/bin/bash

# Copy this file to the EC2 instanse in /usr/local/bin and use it to switch nginx proxying to the desired port (container)

containers=$(docker ps --format "{{.Ports}}|{{.CreatedAt}}|{{.Status}}|{{.Image}}|{{.Names}}")

if [ -z "$containers" ]; then
    echo "No running containers found."
    exit 1
fi

current_port=$(grep -oP "proxy_pass http:\/\/localhost:\K\d+" /etc/nginx/nginx.conf)
if [ -n "$current_port" ]; then
    echo "Current port in Nginx config: $current_port"
else
    echo "No port configured in Nginx config."
fi

echo -e "\nSelect the port from the running containers:"
echo "---------------------------------------------------------"
container_list=()
i=1
while IFS='|' read -r ports created status image name; do
    port=$(echo "$ports" | grep -oP "\d{2,5}(?=\->)" | head -n 1)
    container_list+=("$i) Port: $port | Created: $created | Status: $status | Image: $image | TaskDef: $name")
    i=$((i+1))
done <<< "$containers"

for container in "${container_list[@]}"; do
    echo "$container"
done

echo -n "Enter the number of the container to select its port: "
read selection

if [[ ! "$selection" =~ ^[0-9]+$ ]] || [ "$selection" -lt 1 ] || [ "$selection" -gt "$i" ]; then
    echo "Invalid selection."
    exit 1
fi

selected_container=$(echo "$containers" | sed -n "${selection}p")
selected_port=$(echo "$selected_container" | cut -d '|' -f1 | grep -oP "\d{2,5}(?=\->)" | head -n 1) # Получаем только порт
selected_name=$(echo "$selected_container" | cut -d '|' -f5)
selected_image=$(echo "$selected_container" | cut -d '|' -f4)
selected_status=$(echo "$selected_container" | cut -d '|' -f3)
selected_created=$(echo "$selected_container" | cut -d '|' -f2)

echo -e "\nYou selected the container:"
echo "Port: $selected_port"
echo "Created: $selected_created"
echo "Status: $selected_status"
echo "Image: $selected_image"
echo "Name: $selected_name"

sudo sed -i "s/proxy_pass http:\/\/localhost:[0-9]*;/proxy_pass http:\/\/localhost:$selected_port;/g" /etc/nginx/nginx.conf

sudo systemctl restart nginx
echo "Nginx restarted with port $selected_port."
