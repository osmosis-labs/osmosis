<!--
order: 3
-->

# Clients

## Command Line Interface (CLI)

The CLI has been updated with new queries and transactions for the `x/cron` module. View the entire list below.

### Queries

| Command                   | Subcommand | Arguments | Description                    |
| :------------------------ | :--------- | :-------- | :----------------------------- |
| `osmosisd query cron` | `params`   |           | Get Cron params                |
| `osmosisd query cron` | `crons`    |           | Get the list of the cronJobs   |
| `osmosisd query cron` | `cron`     | [id]      | Get the details of the cronJob |

### Transactions

| Command                | Subcommand        | Arguments                                          | Description                               |
| :--------------------- | :---------------- | :------------------------------------------------- | :---------------------------------------- |
| `osmosisd tx cron` | `register-cron`   | [name] [description] [contract_address] [json_msg] | Register the cron job                     |
| `osmosisd tx cron` | `update-cron-job` | [id] [contract_address] [json_msg]                 | update the cron job                       |
| `osmosisd tx cron` | `delete-cron-job` | [id] [contract_address]                            | delete the cron job for the contract      |
| `osmosisd tx cron` | `toggle-cron-job` | [id]                                               | Toggle the cron job for the given cron ID |
