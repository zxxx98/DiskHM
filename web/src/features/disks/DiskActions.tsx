import { Clock3, MoonStar, RefreshCw } from 'lucide-react';

type DiskActionsProps = {
  unsupported: boolean;
};

export function DiskActions({ unsupported }: DiskActionsProps) {
  return (
    <div className="disk-actions" role="group" aria-label="Disk actions">
      <button className="disk-action-button" type="button" disabled={unsupported}>
        <MoonStar size={14} strokeWidth={2} />
        <span>Sleep now</span>
      </button>
      <button className="disk-action-button" type="button" disabled={unsupported}>
        <Clock3 size={14} strokeWidth={2} />
        <span>Sleep in 30m</span>
      </button>
      <button className="disk-action-button disk-action-button--wake" type="button">
        <RefreshCw size={14} strokeWidth={2} />
        <span>Refresh (wake disk)</span>
      </button>
    </div>
  );
}
