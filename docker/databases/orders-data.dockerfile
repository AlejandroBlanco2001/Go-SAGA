FROM postgres:17.5

ENV POSTGRES_USER=myuser
ENV POSTGRES_PASSWORD=somerandompassword
ENV POSTGRES_DB=orders_database

# For easy portforwarding and connection between microservices
EXPOSE 5432

CMD ["postgres"]