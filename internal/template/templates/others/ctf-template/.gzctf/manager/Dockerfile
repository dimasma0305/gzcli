FROM ubuntu:noble

COPY docker-install.sh .
RUN bash docker-install.sh

RUN curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/install.sh | bash && echo "update version 1.24.0"

COPY run.sh .

CMD ["bash", "./run.sh"]
