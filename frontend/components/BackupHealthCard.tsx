import React from 'react';
import { ShieldCheck, ShieldAlert, Calendar, Settings, AlertTriangle, CheckCircle2 } from 'lucide-react';
import { format, isToday, isTomorrow } from 'date-fns';
import { InstanceMetric, InstanceBackupInfo } from '@/types';

interface BackupHealthCardProps {
  instance: InstanceMetric;
  onConfigureBackup: () => void;
}

interface LastRunStatusProps {
    backupInfo: InstanceBackupInfo;
}

const LastRunStatus: React.FC<LastRunStatusProps> = ({ backupInfo }) => {
    if (!backupInfo.last_status) {
      return <span className="text-zinc-400">Never</span>;
    }

    const status = backupInfo.last_status.toUpperCase();
    
    if (status === 'COMPLETED') {
        return <span className="flex items-center gap-1.5 text-emerald-400"><CheckCircle2 size={14}/> {status}</span>;
    }
    if (status === 'FAILED') {
        return <span className="flex items-center gap-1.5 text-red-400"><AlertTriangle size={14}/> {status}</span>;
    }
    
    return <span className="text-yellow-400">{status}</span>;
};

const BackupHealthCard: React.FC<BackupHealthCardProps> = ({ 
  instance, 
  onConfigureBackup 
}) => {
  const { backup_info } = instance;

  // If backup_info is not available, show not protected state
  if (!backup_info) {
    return (
      <div className="bg-zinc-900 border border-zinc-800 rounded-lg p-4">
        <div className="flex items-center gap-2 mb-4">
          <ShieldAlert className="text-zinc-500" size={20} />
          <h3 className="font-semibold text-zinc-100">Data Protection</h3>
        </div>
        <div className="text-center py-6">
          <p className="text-zinc-400 mb-4">This instance is not protected by automatic backups.</p>
          <button
            onClick={onConfigureBackup}
            className="px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-md flex items-center gap-2 mx-auto transition-colors"
          >
            <Settings size={16} />
            Configure Backup
          </button>
        </div>
      </div>
    );
  }

  const formatNextRun = (nextRun: string) => {
    const date = new Date(nextRun);
    if (isToday(date)) {
      return `Today at ${format(date, 'HH:mm')}`;
    }
    if (isTomorrow(date)) {
      return `Tomorrow at ${format(date, 'HH:mm')}`;
    }
    return format(date, 'MMM d, yyyy \'at\' HH:mm');
  };

  return (
    <div className="bg-zinc-900/50 border border-zinc-800 rounded-lg p-5 shadow-md">
      <div className="flex justify-between items-center mb-4">
        <div className="flex items-center gap-3">
          {backup_info.enabled ? (
            <ShieldCheck className="text-emerald-400" size={22} />
          ) : (
            <ShieldAlert className="text-yellow-500" size={22} />
          )}
          <h3 className="font-semibold text-zinc-100 text-lg">Data Protection</h3>
        </div>
         <button
            onClick={onConfigureBackup}
            className="px-3 py-1.5 bg-zinc-800 hover:bg-zinc-700/80 text-zinc-300 rounded-md flex items-center gap-2 text-xs transition-colors border border-zinc-700"
          >
            <Settings size={14} />
            Configure
        </button>
      </div>
      
      <div className="space-y-3.5 text-sm">
        <div className="flex items-center justify-between">
          <span className="text-zinc-400">Status</span>
          {backup_info.enabled ? (
            <span className="font-medium text-emerald-400">Active</span>
          ) : (
            <span className="font-medium text-zinc-400">Disabled</span>
          )}
        </div>

        {backup_info.enabled && (
          <>
            <div className="flex items-center justify-between">
              <span className="text-zinc-400">Next Run</span>
              <span className="font-medium text-zinc-200 flex items-center gap-2">
                <Calendar size={14} />
                {backup_info.next_run ? formatNextRun(backup_info.next_run) : 'N/A'}
              </span>
            </div>
            
            <div className="flex items-center justify-between">
              <span className="text-zinc-400">Last Run</span>
              <div className="font-medium text-zinc-200">
                <LastRunStatus backupInfo={backup_info} />
              </div>
            </div>
          </>
        )}
      </div>
       {!backup_info.enabled && (
         <div className="text-center mt-5 pt-4 border-t border-zinc-800">
            <p className="text-zinc-500 text-xs">Automatic backups are disabled for this instance.</p>
        </div>
      )}
    </div>
  );
};

export default BackupHealthCard;