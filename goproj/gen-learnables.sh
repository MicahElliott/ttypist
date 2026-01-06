#! /bin/zsh

sfile=words-populate.sql

# Remove head to do whole set
head -120 ../10k-3.num |
     sed -r -e '1d' -e "/'/d" \
         -e "s/^/INSERT INTO learnable (lrank, diffy, lname, course) VALUES ('/" \
         -e "s/\t/', '/g" -e "s/$/', 'typing');/" >! $sfile

print "\nINSERT INTO question ( qtype, lid ) SELECT 'typed', lid from learnable;" >> $sfile

print "Created sql file to populate learnables and questions for typed words: $sfile"
