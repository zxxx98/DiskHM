import type { DiskListItem } from './api';
import { DiskActions } from './DiskActions';

type DiskRowProps = {
  disk: DiskListItem;
};

export function DiskRow({ disk }: DiskRowProps) {
  return (
    <tr>
      <td>
        <div className="disk-table__disk">
          <span className="disk-table__name">{disk.name}</span>
          <span className="disk-table__model">{disk.model}</span>
        </div>
      </td>
      <td>
        <span className="disk-table__pill">{disk.powerState}</span>
      </td>
      <td>
        <span className="disk-table__freshness">{disk.refreshFreshness}</span>
      </td>
      <td>
        <DiskActions unsupported={disk.unsupported} />
      </td>
    </tr>
  );
}
