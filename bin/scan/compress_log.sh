#!/bin/bash
if [ "$1" == "-h" ]; then
  echo ""
  echo "Compress yesterday log"
  echo "Usage: `basename LogDirPath"
  echo "e.g:"
  echo "       `basename ~/saveio/scan/Log"
  echo ""
  exit 0
fi

cd $1
tar -zcvf $1/`date -d "-1 week" +%F`.tar.gz *`date -d "-1 week" +%F`*.log
rm -rf *`date -d "-1 week" +%F`*.log
echo "compress `date -d "-1 week" +%F` log done"