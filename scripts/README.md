# LlamaChat Scripts

This directory contains utility scripts for development, deployment, and management of the LlamaChat application.

## Available Scripts

### Database Scripts

- `init-db.sql` - SQL script to initialize the database schema and initial data
- `migrate.sh` - Shell script to run database migrations
- `backup-db.sh` - Shell script to backup the database

### Deployment Scripts

- `deploy.sh` - Script to deploy the application to production
- `build.sh` - Script to build the application for production
- `release.sh` - Script to create a new release

### Development Scripts

- `setup-dev.sh` - Script to set up the development environment
- `lint.sh` - Script to run linters on the codebase
- `test.sh` - Script to run tests
- `generate-mock.sh` - Script to generate mock data

### Utility Scripts

- `encrypt-config.sh` - Script to encrypt sensitive configuration
- `decrypt-config.sh` - Script to decrypt sensitive configuration
- `generate-jwt-secret.sh` - Script to generate a secure JWT secret
- `install-hooks.sh` - Script to install Git hooks

## Usage

Most scripts can be run directly:

```bash
./scripts/script-name.sh [arguments]
```

Some scripts may require specific arguments or environment variables. Check the comments at the top of each script for specific usage instructions.

## Creating New Scripts

When adding a new script to this directory:

1. Make sure it's executable (`chmod +x script-name.sh`)
2. Add a comment header explaining the purpose and usage
3. Add proper error handling and logging
4. Update this README with information about the new script

## Script Standards

- All scripts should have a help/usage option (-h or --help)
- Scripts should return appropriate exit codes (0 for success, non-zero for errors)
- Scripts should use proper logging to indicate progress and errors
- Use shellcheck to validate shell scripts 