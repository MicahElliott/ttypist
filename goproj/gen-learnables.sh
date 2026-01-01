#! /bin/zsh

# Remove head to do whole set
head -120 ../10k-3.num |
     sed -r -e '1d' -e "/'/d" \
         -e "s/^/INSERT INTO learnable (lrank, diffy, lname, course) VALUES ('/" \
         -e "s/\t/', '/g" -e "s/$/', 'typing');/" >! t1-schema.sql
