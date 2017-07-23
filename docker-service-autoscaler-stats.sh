ps --no-headers -eo pcpu,pmem,rss \
| gawk 'BEGIN{cpu=0;mem=0;rss=0} {cpu=cpu+$1;mem=mem+$2;rss=rss+$3} END{print "{\"cpu\":"cpu",\"mem\":"mem",\"rss\":"rss"}"}'
