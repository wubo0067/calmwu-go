sub_dirs = indexsvr_main \
           logsvr_main \
           guidesvr_main \
           csssvr_main \
		   financesvr_main \
		   fleetsvr_main

.PHONY: clean all
all:
	@for subdir in $(sub_dirs); \
	do \
		if ! [ -d $$subdir ]; then echo "[$$subdir] is not exist!"; exit -1; fi; \
		make -C $$subdir; \
		if [ $$? -ne 0 ]; then echo "build [$$subdir] failed"; exit -1; else echo "build [$$subdir] successed"; fi \
	done

clean:
	@(for i in $(sub_dirs); do ( cd $$i && make clean ); done)


