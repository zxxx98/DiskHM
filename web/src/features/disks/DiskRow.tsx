import type { DiskListItem } from './api';
import { DiskActions } from './DiskActions';

type DiskRowProps = {
  disk: DiskListItem;
  onSleepNow?: (diskID: string) => void;
  onSleepLater?: (diskID: string) => void;
  onWakeRefresh?: (diskID: string) => void;
};

export function DiskRow({ disk, onSleepNow, onSleepLater, onWakeRefresh }: DiskRowProps) {
  const disableSleepActions = disk.unsupported;

  return (
    <tr>
      <td>
        <div className="disk-table__disk">
          <span className="disk-table__name">{disk.name}</span>
          <span className="disk-table__model">{disk.model}</span>
          {disk.mounts?.length ? <span className="disk-table__freshness">{disk.mounts.join(', ')}</span> : null}
        </div>
      </td>
      <td>
        <span className="disk-table__pill">{disk.powerState}</span>
      </td>
      <td>
        <span className="disk-table__freshness">{disk.refreshFreshness}</span>
      </td>
      <td>
        <DiskActions
          disableSleepNow={disableSleepActions}
          disableSleepLater={disableSleepActions}
          disableWakeRefresh={false}
          onSleepNow={() => onSleepNow?.(disk.id)}
          onSleepLater={() => onSleepLater?.(disk.id)}
          onWakeRefresh={() => onWakeRefresh?.(disk.id)}
        />
      </td>
    </tr>
  );
}
