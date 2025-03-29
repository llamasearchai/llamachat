# Contributing to LlamaChat

Thank you for considering contributing to LlamaChat! This document outlines the process for contributing to the project.

## Code of Conduct

By participating in this project, you agree to abide by its Code of Conduct. Please be respectful and considerate of others.

## How Can I Contribute?

### Reporting Bugs

- Before submitting a bug report, please check if the issue has already been reported.
- Use the bug report template to create a new issue.
- Provide as much information as possible to help reproduce and fix the bug.

### Suggesting Enhancements

- Before submitting an enhancement suggestion, please check if it has already been suggested.
- Use the feature request template to create a new issue.
- Provide a clear rationale for why this enhancement would be valuable.

### Pull Requests

1. Fork the repository
2. Create a branch for your feature or bugfix (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Development Setup

1. Install Go 1.21 or higher
2. Set up PostgreSQL 12 or higher
3. Clone the repository
4. Run `make db-setup` to set up the database
5. Run `make run` to start the application

## Coding Standards

- Follow Go best practices
- Use `gofmt` for formatting
- Add tests for new code
- Document all exported functions, types, and methods
- Follow the existing code structure and patterns

## Testing

Run tests with:

```bash
make test
```

For test coverage:

```bash
make test-coverage
```

## Documentation

Please update documentation when changing functionality or adding new features.

## Community

If you have questions, feel free to open an issue for discussion.

Thank you for contributing to LlamaChat! 