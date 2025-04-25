FROM nginx:stable-alpine

# Копируем конфиги
COPY deployments/nginx/nginx.conf /etc/nginx/nginx.conf
COPY deployments/nginx/conf.d /etc/nginx/conf.d

COPY ../website/ /usr/share/nginx/html

# Устанавливаем grpc-web модуль
RUN apk add --no-cache nginx-mod-http-grpc

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]