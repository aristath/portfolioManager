"""ML job task functions."""

import logging

logger = logging.getLogger(__name__)


async def analytics_regime(detector, monolith) -> None:
    securities = await monolith.get_securities(active_only=True, ml_enabled_only=False)
    symbols = [s["symbol"] for s in securities]
    if len(symbols) < 3:
        logger.warning("Not enough securities for regime detection")
        return

    model = await detector.train_model(symbols)
    if model:
        logger.info("Regime model trained on %d securities", len(symbols))


async def ml_retrain(monolith, retrainer) -> None:
    securities = await monolith.get_ml_enabled_securities()

    if not securities:
        logger.info("No ML-enabled securities to retrain")
        return

    trained = 0
    skipped = 0

    for sec in securities:
        symbol = sec["symbol"]
        result = await retrainer.retrain_symbol(symbol)

        if result:
            logger.info(
                "ML retraining complete for %s: RMSE=%.4f, samples=%s",
                symbol,
                result.get("validation_rmse", 0),
                result.get("training_samples", 0),
            )
            trained += 1
        else:
            logger.info("ML retraining skipped for %s: insufficient data", symbol)
            skipped += 1

    logger.info("ML retraining complete: %d trained, %d skipped", trained, skipped)


async def ml_monitor(monolith, monitor) -> None:
    securities = await monolith.get_ml_enabled_securities()

    if not securities:
        logger.info("No ML-enabled securities to monitor")
        return

    monitored = 0

    for sec in securities:
        symbol = sec["symbol"]
        per_model = await monitor.track_symbol_performance(symbol)

        if not per_model:
            logger.info("ML monitoring for %s: no predictions to evaluate", symbol)
            continue

        for mt, metrics in per_model.items():
            if metrics.get("predictions_evaluated", 0) > 0:
                logger.info(
                    "ML performance for %s/%s: MAE=%.4f, RMSE=%.4f, N=%d",
                    symbol,
                    mt,
                    metrics.get("mean_absolute_error", 0),
                    metrics.get("root_mean_squared_error", 0),
                    metrics["predictions_evaluated"],
                )
                if metrics.get("drift_detected"):
                    logger.warning("ML DRIFT DETECTED for %s/%s", symbol, mt)

        monitored += 1

    logger.info("ML monitoring complete: %d securities evaluated", monitored)
