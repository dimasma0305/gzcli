FROM ubuntu:noble

COPY docker-install.sh .
RUN bash docker-install.sh

RUN curl -s https://raw.githubusercontent.com/dimasma0305/gzcli/main/install.sh | bash && echo "update version 1.17.1"

COPY run.sh .

CMD ["bash", "./run.sh"]