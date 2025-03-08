# Contributing to Event-Driven Library

Thank you for your interest in contributing to our event-driven library! This document provides guidelines for contributing to this project.

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Create a new branch from `main`
4. Make your changes
5. Push to your fork
6. Open a Pull Request

## Development Requirements

- Go 1.20 or higher
- PostgreSQL
- Docker (optional, for local development)

## Pull Request Process

1. Create a branch from `main` using the following naming convention:
    - Feature: `feature/your-feature-name`
    - Bug Fix: `fix/issue-description`
    - Documentation: `docs/what-changed`

2. Ensure your code:
    - Has appropriate test coverage
    - Follows Go coding standards
    - Includes documentation updates if needed
    - Passes all existing tests

3. Update the README.md if necessary

## Code Guidelines

- Write clear, documented code
- Follow Go best practices and idioms
- Use meaningful variable and function names
- Add comments for complex logic
- Include unit tests for new features

## Database Changes

When making database changes:
- Include SQL migration scripts
- Update the SQL queries in `internal/sqlc/`
- Test migrations both up and down

## Commit Messages

Use clear and meaningful commit messages: