import type { DiskListItem } from './api';
import { DiskRow } from './DiskRow';

type DiskTablePageProps = {
  disks: DiskListItem[];
  actionError?: string | null;
  onSleepNow?: (diskID: string) => void;
  onSleepLater?: (diskID: string) => void;
  onWakeRefresh?: (diskID: string) => void;
};

export function DiskTablePage({ disks, actionError, onSleepNow, onSleepLater, onWakeRefresh }: DiskTablePageProps) {
  return (
    <main className="disk-page">
      <section className="disk-page__panel" aria-labelledby="disk-page-title">
        <div className="disk-page__header">
          <div>
            <p className="disk-page__eyebrow">Disk inventory</p>
            <h1 id="disk-page-title">Disks</h1>
          </div>
          <p className="disk-page__copy">
            Safe metadata stays cached until you request a fresh probe. Wake-capable refresh is explicit so sleeping
            disks are only touched on purpose.
          </p>
        </div>
        {actionError ? <p role="alert">{actionError}</p> : null}

        <div className="disk-table-wrap">
          <table className="disk-table">
            <thead>
              <tr>
                <th scope="col">Disk</th>
                <th scope="col">State</th>
                <th scope="col">Freshness</th>
                <th scope="col">Actions</th>
              </tr>
            </thead>
            <tbody>
              {disks.map((disk) => (
                <DiskRow
                  key={disk.id}
                  disk={disk}
                  onSleepLater={onSleepLater}
                  onSleepNow={onSleepNow}
                  onWakeRefresh={onWakeRefresh}
                />
              ))}
            </tbody>
          </table>
        </div>
      </section>
    </main>
  );
}
