FROM nginx:alpine

# 复制构建产物到 nginx 目录
COPY ./dist /usr/share/nginx/html

# 暴露端口
EXPOSE 3000

# 启动 nginx
CMD ["nginx", "-g", "daemon off;"]

