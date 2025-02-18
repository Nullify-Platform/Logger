from pylogtracer import structured_logger, track

# Basic logging
structured_logger.info("Application started", environment="production", version="1.0.0")
structured_logger.debug("Debug message", user_id="123")
structured_logger.error("Something went wrong", error_code=500)


def main():
    # Basic logging
    structured_logger.info(
        "Application started", environment="production", version="1.0.0"
    )
    structured_logger.debug("Debug message", user_id="123")
    structured_logger.error("Something went wrong", error_code=500)
    result = process_order("ORDER123", "USER456")
    return result


@track(span_name="process_order")
def process_order(order_id, user_id):
    structured_logger.info("Processing order", order_id=order_id, user_id=user_id)
    return process_order1(order_id, user_id)


@track(span_name="process_order1")
def process_order1(order_id, user_id):
    structured_logger.info("Processing order 1", order_id=order_id, user_id=user_id)
    return {"status": "success"}


# # Call the traced function
# result = process_order("ORDER123", "USER456")
if __name__ == "__main__":
    main()
