#!/usr/bin/env python3
"""Validate REST service implementations."""

import importlib
import sys
from pathlib import Path

# Add project root to path
sys.path.insert(0, str(Path(__file__).parent))


def validate_service(service_name: str) -> tuple[bool, str]:
    """
    Validate a service can be imported and has required components.

    Args:
        service_name: Name of the service (e.g., "universe")

    Returns:
        Tuple of (success, message)
    """
    try:
        # Import service modules
        models_module = f"services.{service_name}.models"
        routes_module = f"services.{service_name}.routes"
        deps_module = f"services.{service_name}.dependencies"
        main_module = f"services.{service_name}.main"

        # Try importing each module
        importlib.import_module(models_module)
        importlib.import_module(routes_module)
        importlib.import_module(deps_module)
        main = importlib.import_module(main_module)

        # Verify FastAPI app exists
        if not hasattr(main, "app"):
            return False, f"❌ {service_name}: No 'app' attribute in main module"

        return True, f"✅ {service_name}: All imports successful, FastAPI app found"

    except ImportError as e:
        return False, f"❌ {service_name}: Import error - {e}"
    except Exception as e:
        return False, f"❌ {service_name}: Unexpected error - {e}"


def validate_http_client(service_name: str) -> tuple[bool, str]:
    """
    Validate HTTP client exists for a service.

    Args:
        service_name: Name of the service

    Returns:
        Tuple of (success, message)
    """
    try:
        client_module = f"app.infrastructure.http_clients.{service_name}_client"
        importlib.import_module(client_module)
        return True, f"✅ {service_name}_client: HTTP client available"
    except ImportError as e:
        return False, f"❌ {service_name}_client: Import error - {e}"
    except Exception as e:
        return False, f"❌ {service_name}_client: Unexpected error - {e}"


def main():
    """Run validation for all services."""
    print("=" * 70)
    print("REST Service Validation")
    print("=" * 70)

    services = [
        "universe",
        "portfolio",
        "trading",
        "scoring",
        "optimization",
        "planning",
        "gateway",
    ]

    print("\n📦 Validating Service Implementations...")
    print("-" * 70)

    service_results = []
    for service in services:
        success, message = validate_service(service)
        service_results.append(success)
        print(message)

    print("\n🔗 Validating HTTP Clients...")
    print("-" * 70)

    client_results = []
    for service in services:
        success, message = validate_http_client(service)
        client_results.append(success)
        print(message)

    print("\n🔍 Validating Service Locator...")
    print("-" * 70)

    try:
        from app.infrastructure.service_discovery import get_service_locator

        locator = get_service_locator()

        # Test HTTP client creation
        test_results = []
        for service in services:
            try:
                client = locator.create_http_client(service)
                print(f"✅ ServiceLocator.create_http_client('{service}'): Success")
                test_results.append(True)
            except Exception as e:
                print(f"❌ ServiceLocator.create_http_client('{service}'): {e}")
                test_results.append(False)

        locator_success = all(test_results)
    except Exception as e:
        print(f"❌ Service locator error: {e}")
        locator_success = False

    print("\n" + "=" * 70)
    print("Summary")
    print("=" * 70)

    total_services = len(services)
    services_passed = sum(service_results)
    clients_passed = sum(client_results)

    print(f"✅ Services validated: {services_passed}/{total_services}")
    print(f"✅ HTTP clients validated: {clients_passed}/{total_services}")
    print(f"✅ Service locator: {'PASS' if locator_success else 'FAIL'}")

    all_passed = (
        services_passed == total_services
        and clients_passed == total_services
        and locator_success
    )

    if all_passed:
        print("\n🎉 ALL VALIDATIONS PASSED!")
        return 0
    else:
        print("\n⚠️  Some validations failed. Please review errors above.")
        return 1


if __name__ == "__main__":
    sys.exit(main())
