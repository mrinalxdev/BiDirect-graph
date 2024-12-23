# BiDirect - Social Graph Service

BiDirect is a minimalist implementation of a distributed social graph service, designed to demonstrate the core concepts of building scalable social networking systems like LinkedIn or Facebook's social graph infrastructure.

## Overview

This implementation is a micro-instance of how large-scale social networks manage connection data and calculate relationship distances. While production systems handle billions of connections across thousands of nodes, this implementation uses a smaller scale to demonstrate the key architectural concepts.

### Key Features
- Distributed graph storage using Redis
- Partitioned data architecture
- Connection degree calculation (1st, 2nd, 3rd degree connections)
- Shared connection discovery
- Network distance computation

## Architecture

### Scaled-Down Components

1. **Partitioning Strategy**
   - Current: Simple modulo-based partitioning across 3 Redis nodes
   - Production: Would use consistent hashing or range-based sharding across thousands of nodes

2. **Caching Layer**
   - Current: Single Redis instance for second-degree connections
   - Production: Multi-tiered caching with in-memory, near-memory, and disk-based caches

3. **Node Management**
   - Current: Static node configuration
   - Production: Dynamic node discovery and automated rebalancing

### API Endpoints

```
GET  /api/connections/{memberID}           # Get member's direct connections
GET  /api/shared-connections/{id1}/{id2}   # Find shared connections
POST /api/distances                        # Calculate network distances
```

## Technical Design Decisions

### Data Distribution
- Uses partition-based distribution to demonstrate how social graphs can be split across multiple nodes
- Each node manages multiple partitions to show how load can be distributed
- Simplified partitioning function for demonstration purposes

### Connection Traversal
- Implements efficient 2nd and 3rd-degree connection discovery
- Uses Redis sorted sets for quick connection lookups
- Demonstrates caching of frequently accessed paths

### Set Cover Algorithm
- Implements a greedy approach for finding minimum node sets
- Shows how to optimize multi-node queries in a distributed system

## Getting Started

### Prerequisites
- Go 1.21+
- Docker and Docker Compose

### Running the Service

1. Start the infrastructure:
```bash
docker-compose up -d
```

2. The service will be available at `http://localhost:8080`

3. Load sample data:
```bash
python imp.py
```

## Production Considerations

This implementation is intentionally simplified. In a production environment, you would need to consider:

1. **Scalability**
   - Current: 3 Redis nodes, 10 partitions per node
   - Production: Thousands of nodes, dynamic partition allocation

2. **Reliability**
   - Current: Basic error handling
   - Production: Circuit breakers, fallbacks, redundancy

3. **Performance**
   - Current: Simple caching strategy
   - Production: Multi-level caching, pre-computation of common paths

4. **Monitoring**
   - Current: Basic logging
   - Production: Comprehensive metrics, tracing, alerting

5. **Security**
   - Current: No authentication
   - Production: OAuth, rate limiting, encryption

## Design Choices

### Why Redis?
- Demonstrates in-memory graph storage principles
- Sorted sets provide efficient connection lookups
- Easy to understand and set up for demonstration

### Why Partition-Based Distribution?
- Shows basic concepts of data sharding
- Demonstrates how to handle cross-partition queries
- Simplified version of production-grade distribution strategies

## Contributing

This is an educational project designed to demonstrate distributed systems concepts. Contributions that help clarify these concepts or add new educational examples are welcome.

## License

MIT License

## Acknowledgments

This implementation draws inspiration from real-world social graph systems while maintaining a focus on educational value and clarity over production-grade features.