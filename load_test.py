#!/usr/bin/env python3
import json
import time
import subprocess
import sys
import argparse
from concurrent.futures import ThreadPoolExecutor, as_completed
from tqdm import tqdm

def load_queries(file_path):
    with open(file_path, 'r', encoding='utf-8') as f:
        data = json.load(f)
    return [item['query'] for item in data]

def make_request(query):
    start = time.perf_counter()
    try:
        result = subprocess.run(
            ['grpcurl', '-plaintext', '-d', json.dumps({"question": query}), 
             'localhost:8083', 'gateway.GatewayService/Ask'],
            capture_output=True, text=True, timeout=180
        )
        elapsed = time.perf_counter() - start
        success = result.returncode == 0
        err = result.stderr if not success else None
        if err:
            print(f"  → {err[:150]}", file=sys.stderr)
        return elapsed, success, err
    except subprocess.TimeoutExpired:
        elapsed = time.perf_counter() - start
        return elapsed, False, "timeout"
    except Exception as e:
        elapsed = time.perf_counter() - start
        return elapsed, False, str(e)

def run_load_test(queries, concurrency, total_requests):
    results = []
    
    with ThreadPoolExecutor(max_workers=concurrency) as executor:
        futures = {}
        for i in range(total_requests):
            query = queries[i % len(queries)]
            future = executor.submit(make_request, query)
            futures[future] = i
        
        with tqdm(total=total_requests, desc="Requests") as pbar:
            for future in as_completed(futures):
                results.append(future.result())
                pbar.update(1)
    
    return results

def print_stats(results):
    times = [r[0] for r in results if r[1]]
    errors = sum(1 for r in results if not r[1])
    
    if not times:
        print("No successful requests")
        return
    
    times.sort()
    total = len(times)
    p50 = times[int(total * 0.5)]
    p95 = times[int(total * 0.95)]
    p99 = times[int(total * 0.99)]
    avg = sum(times) / total
    
    print(f"\n=== Results ===")
    print(f"Total requests: {len(results)}")
    print(f"Successful: {total}")
    print(f"Errors: {errors}")
    print(f"\nLatency (seconds):")
    print(f"  Avg: {avg:.3f}s")
    print(f"  p50: {p50:.3f}s")
    print(f"  p95: {p95:.3f}s")
    print(f"  p99: {p99:.3f}s")
    print(f"  Min: {min(times):.3f}s")
    print(f"  Max: {max(times):.3f}s")
    
    if errors > 0:
        print(f"\nErrors: {errors}/{len(results)} ({100*errors/len(results):.1f}%)")
        error_msgs = [r[2] for r in results if not r[1] and r[2]]
        for e in set(error_msgs):
            print(f"  - {e[:100]}")

def main():
    parser = argparse.ArgumentParser(description='gRPC Load Test for RAG Gateway')
    parser.add_argument('--dataset', default='rag/rag_eval_dataset_1.json', help='Path to queries JSON')
    parser.add_argument('-c', '--concurrency', type=int, default=1, help='Concurrent connections')
    parser.add_argument('-n', '--requests', type=int, default=10, help='Total requests')
    args = parser.parse_args()
    
    queries = load_queries(args.dataset)
    print(f"Loaded {len(queries)} queries from {args.dataset}")
    
    print(f"Starting test: {args.requests} requests, {args.concurrency} concurrent...")
    results = run_load_test(queries, args.concurrency, args.requests)
    print_stats(results)

if __name__ == '__main__':
    main()