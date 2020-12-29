#! /bin/bash

objfiles=`awk '/\.o/{print $1}' .depend`
#echo $objfiles

for objfile in $objfiles
do
	dest_objfile="obj\/"$objfile
	#echo $objfile
	#echo $dest_objfile
	sed 's/'"$objfile"'/'"$dest_objfile"'/g' .depend > .depend.temp
	mv .depend.temp .depend
done

