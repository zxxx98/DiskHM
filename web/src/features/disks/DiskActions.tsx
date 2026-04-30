import { Clock3, MoonStar, RefreshCw } from 'lucide-react';

type DiskActionsProps = {
  disableSleepNow: boolean;
  disableSleepLater: boolean;
  disableWakeRefresh?: boolean;
  onSleepNow?: () => void;
  onSleepLater?: () => void;
  onWakeRefresh?: () => void;
};

export function DiskActions({
  disableSleepNow,
  disableSleepLater,
  disableWakeRefresh = false,
  onSleepNow,
  onSleepLater,
  onWakeRefresh,
}: DiskActionsProps) {
  return (
    <div className="disk-actions" role="group" aria-label="Disk actions">
      <button className="disk-action-button" type="button" disabled={disableSleepNow} onClick={onSleepNow}>
        <MoonStar size={14} strokeWidth={2} />
        <span>Sleep now</span>
      </button>
      <button className="disk-action-button" type="button" disabled={disableSleepLater} onClick={onSleepLater}>
        <Clock3 size={14} strokeWidth={2} />
        <span>Sleep in 30m</span>
      </button>
      <button
        className="disk-action-button disk-action-button--wake"
        type="button"
        disabled={disableWakeRefresh}
        onClick={onWakeRefresh}
      >
        <RefreshCw size={14} strokeWidth={2} />
        <span>Refresh (wake disk)</span>
      </button>
    </div>
  );
}
