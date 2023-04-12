docker-compose up -d

## Notes
The solution is far from optimal due to time constraints.

- All three applications are single files without proper package organisation
- Tests are missing
- Duplicated rabbitmq and redis connections
- Would usually apply a kind of ports & adapters architecture style in order to abstract the right concepts
- etc ...
