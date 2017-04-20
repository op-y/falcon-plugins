BEGIN {
    FS=":"
}

/KEYWORD0/ {
    if($NF>4000) {
        print "RPUSH count.KEYWORD0 "$NF
    }
}

/KEYWORD1/ {
    if($NF>4000) {
        print "RPUSH count.KEYWORD1 "$NF
    };
}

/KEYWORD2/ {
    if($NF>4000) {
        print "RPUSH count.KEYWORD2 "$NF
    }
}
