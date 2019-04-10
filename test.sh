#/bin/bash
go run main.go pull.go pull -o out.exe -t 10 -w .tmp --repo https://github.com/untillpro/directcd-test \
--replace https://github.com/untillpro/directcd-test-print=https://github.com/untillpro/directcd-test-print -- -o1 arg1 arg2
#--replace https://github.com/untillpro/directcd-test-print -- -o1 arg1 arg2


