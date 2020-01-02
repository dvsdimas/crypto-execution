# This project was designed and developed by me for MSQ.AI firm, but they've decided to take my work, ditch me out and not to pay for it. In this case I have all rights to keep this project as my property and make it open source under MIT license.


Main idea of the project is providing plain REST api for trading in any possible places such as: Crypto Exchanges, Brokers, whatever that allows to buy and sell. Also all execution steps are persisted in Postgress and after any software failures, system consistency will be restored, all uncetanties in execution states will be clarified at trading providers with theirs API automatically during recovery procedure.

It designed with idea that start up should be cheap, but with possibility to raise and handle big load without additional afforts. 

So it is just execution and the logic for initiation the trade, choosing a sourse for trade, instrument, amount and price are out of the project.


# msq-execution REST API


# PUT MARKET BUY 

curl -X PUT -d "cmd[exchange]=BINANCE&cmd[instrument]=BTTBTC&cmd[direction]=BUY&cmd[order_type]=MARKET&cmd[time_in_force]=GTC&cmd[amount]=10000&cmd[execution_type]=OPEN&cmd[account_id]=1&cmd[api_key]=JbOlqQXlxPGP&cmd[secret_key]=xfPip87DozwjXe6&cmd[finger_print]=unique_id" localhost:8080/execution/v1/command/


# PUT LIMIT BUY 

curl -X PUT -d "cmd[exchange]=BINANCE&cmd[instrument]=BTTBTC&cmd[direction]=BUY&cmd[order_type]=LIMIT&cmd[limit_price]=0.00000005&cmd[time_in_force]=FOK&cmd[amount]=10000&cmd[execution_type]=OPEN&cmd[account_id]=1&cmd[api_key]=JbOlqQXlxPGPrO8fLk&cmd[secret_key]=xfPip87Dozwj&cmd[finger_print]=unique_id" localhost:8080/execution/v1/command/


# PUT MARKET SELL

curl -X PUT -d "cmd[exchange]=BINANCE&cmd[instrument]=BTTBTC&cmd[direction]=SELL&cmd[order_type]=MARKET&cmd[time_in_force]=FOK&cmd[amount]=10000&cmd[execution_type]=OPEN&cmd[account_id]=1&cmd[api_key]=JbOlqQXlxPGPrO8fLk&cmd[secret_key]=xfPip87DozwjX&cmd[finger_print]=unique_id" localhost:8080/execution/v1/command/


# PUT LIMIT SELL

curl -X PUT -d "cmd[exchange]=BINANCE&cmd[instrument]=BTTBTC&cmd[direction]=SELL&cmd[order_type]=LIMIT&cmd[limit_price]=0.00000006&cmd[time_in_force]=FOK&cmd[amount]=10000&cmd[execution_type]=OPEN&cmd[account_id]=1&cmd[api_key]=JbOlqQXlxPGPr&cmd[secret_key]=xfPip87DozwjXe64&cmd[finger_print]=unique_id" localhost:8080/execution/v1/command/


# PUT INFO

curl -X PUT -d "cmd[exchange]=BINANCE&cmd[instrument]= &cmd[direction]=ACCOUNT&cmd[order_type]=INFO&cmd[time_in_force]=FOK&cmd[amount]=0&cmd[execution_type]=REQUEST&cmd[account_id]=1&cmd[api_key]=JbOlqQXlxPGPrO8fL&cmd[secret_key]=xfPip87DozwjXe64Ci&cmd[finger_print]=unique_id" localhost:8080/execution/v1/command/


# GET STATUS

curl -X GET localhost:8080/execution/v1/command/61
