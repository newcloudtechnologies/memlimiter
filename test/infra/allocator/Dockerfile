# TODO: либо менять базовый образ при смене ОС, либо наследоваться от образа с компилятором и собирать исходники прямо там
FROM fedora:35

ADD ucs-allocator /usr/local/bin

CMD /usr/local/bin/ucs-allocator server -c /etc/ucs/allocator/config.json